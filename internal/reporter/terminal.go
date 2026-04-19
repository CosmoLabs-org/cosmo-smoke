package reporter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	passStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	failStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red
	skipStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
	boldStyle  = lipgloss.NewStyle().Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))  // cyan
)

// Terminal writes colored test output to a writer.
type Terminal struct {
	w io.Writer
}

// NewTerminal creates a terminal reporter writing to w.
func NewTerminal(w io.Writer) *Terminal {
	return &Terminal{w: w}
}

func (t *Terminal) PrereqStart(name string) {
	fmt.Fprintf(t.w, "  %s %s", dimStyle.Render("●"), name)
}

func (t *Terminal) PrereqResult(r PrereqResultData) {
	if r.Passed {
		out := ""
		if r.Output != "" {
			out = dimStyle.Render(" (" + r.Output + ")")
		}
		fmt.Fprintf(t.w, "\r  %s %s%s\n", passStyle.Render("✓"), r.Name, out)
	} else {
		fmt.Fprintf(t.w, "\r  %s %s\n", failStyle.Render("✗"), r.Name)
		if r.Hint != "" {
			fmt.Fprintf(t.w, "    %s %s\n", labelStyle.Render("hint:"), r.Hint)
		}
	}
}

func (t *Terminal) TestStart(name string) {
	fmt.Fprintf(t.w, "  %s %s", dimStyle.Render("●"), name)
}

func (t *Terminal) TestResult(r TestResultData) {
	dur := formatDuration(r.Duration)

	if r.Skipped {
		fmt.Fprintf(t.w, "\r  %s %s %s\n", skipStyle.Render("⊘"), r.Name, dimStyle.Render(dur))
		return
	}

	if r.Passed {
		fmt.Fprintf(t.w, "\r  %s %s %s\n", passStyle.Render("✓"), r.Name, dimStyle.Render(dur))
	} else if r.AllowedFailure {
		fmt.Fprintf(t.w, "\r  %s %s %s\n", skipStyle.Render("~"), r.Name, dimStyle.Render(dur+" allowed"))
		for _, a := range r.Assertions {
			if !a.Passed {
				fmt.Fprintf(t.w, "    %s expected %s, got %s\n",
					skipStyle.Render(a.Type+":"),
					a.Expected,
					a.Actual)
			}
		}
		if r.Error != nil {
			fmt.Fprintf(t.w, "    %s %s\n", skipStyle.Render("error:"), r.Error)
		}
	} else {
		fmt.Fprintf(t.w, "\r  %s %s %s\n", failStyle.Render("✗"), r.Name, dimStyle.Render(dur))
		for _, a := range r.Assertions {
			if !a.Passed {
				fmt.Fprintf(t.w, "    %s expected %s, got %s\n",
					failStyle.Render(a.Type+":"),
					a.Expected,
					a.Actual)
			}
		}
		if r.Error != nil {
			fmt.Fprintf(t.w, "    %s %s\n", failStyle.Render("error:"), r.Error)
		}
	}
}

func (t *Terminal) Summary(s SuiteResultData) {
	fmt.Fprintln(t.w)

	parts := []string{
		fmt.Sprintf("%d tests", s.Total),
	}
	if s.Passed > 0 {
		parts = append(parts, passStyle.Render(fmt.Sprintf("%d passed", s.Passed)))
	}
	if s.Failed > 0 {
		parts = append(parts, failStyle.Render(fmt.Sprintf("%d failed", s.Failed)))
	}
	if s.Skipped > 0 {
		parts = append(parts, skipStyle.Render(fmt.Sprintf("%d skipped", s.Skipped)))
	}
	if s.AllowedFailures > 0 {
		parts = append(parts, dimStyle.Render(fmt.Sprintf("%d allowed-failure", s.AllowedFailures)))
	}
	parts = append(parts, dimStyle.Render(formatDuration(s.Duration)))

	fmt.Fprintf(t.w, "  %s\n", strings.Join(parts, "  "))

	if s.TraceHealthPct > 0 {
		pct := fmt.Sprintf("%.1f%%", s.TraceHealthPct)
		if s.TraceDegraded {
			fmt.Fprintf(t.w, "  %s %s\n", failStyle.Render("trace health:"), failStyle.Render(pct+" degraded"))
		} else {
			fmt.Fprintf(t.w, "  %s %s\n", dimStyle.Render("trace health:"), dimStyle.Render(pct))
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("(%dµs)", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("(%dms)", d.Milliseconds())
	}
	return fmt.Sprintf("(%.1fs)", d.Seconds())
}

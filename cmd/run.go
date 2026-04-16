package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/reporter"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/runner"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run smoke tests",
	Long:  "Execute smoke tests defined in .smoke.yaml",
	RunE:  runSmoke,
}

var (
	configFile  string
	tags        []string
	excludeTags []string
	format      string
	failFast    bool
	timeout     string
	dryRun      bool
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&configFile, "file", "f", ".smoke.yaml", "Config file path")
	runCmd.Flags().StringSliceVar(&tags, "tag", nil, "Include only tests with these tags")
	runCmd.Flags().StringSliceVar(&excludeTags, "exclude-tag", nil, "Exclude tests with these tags")
	runCmd.Flags().StringVar(&format, "format", "terminal", "Output format (terminal|json|junit)")
	runCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop on first failure")
	runCmd.Flags().StringVar(&timeout, "timeout", "", "Per-test timeout override (e.g. 30s)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "List tests without running")
}

func runSmoke(cmd *cobra.Command, args []string) error {
	cfg, err := schema.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := schema.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configDir := filepath.Dir(configFile)
	if !filepath.IsAbs(configDir) {
		cwd, _ := os.Getwd()
		configDir = filepath.Join(cwd, configDir)
	}

	// Create reporter
	var rep reporter.Reporter
	switch format {
	case "json":
		rep = reporter.NewJSON(os.Stdout)
	case "junit":
		rep = reporter.NewJUnit(os.Stdout)
	default:
		rep = reporter.NewTerminal(os.Stdout)
	}

	// Parse timeout
	var timeoutDur time.Duration
	if timeout != "" {
		timeoutDur, err = time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", timeout, err)
		}
	}

	r := &runner.Runner{
		Config:    cfg,
		Reporter:  rep,
		ConfigDir: configDir,
	}

	result, err := r.Run(runner.RunOptions{
		Tags:        tags,
		ExcludeTags: excludeTags,
		FailFast:    failFast,
		DryRun:      dryRun,
		Timeout:     timeoutDur,
	})
	if err != nil {
		return err
	}

	if result.Failed > 0 {
		os.Exit(1)
	}
	return nil
}

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/monorepo"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/reporter"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/runner"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
	"github.com/fsnotify/fsnotify"
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
	watch       bool
	envName     string
	monorepoMode bool
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&configFile, "file", "f", ".smoke.yaml", "Config file path")
	runCmd.Flags().StringSliceVar(&tags, "tag", nil, "Include only tests with these tags")
	runCmd.Flags().StringSliceVar(&excludeTags, "exclude-tag", nil, "Exclude tests with these tags")
	runCmd.Flags().StringVar(&format, "format", "terminal", "Output format (terminal|json|junit|tap|prometheus)")
	runCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop on first failure")
	runCmd.Flags().StringVar(&timeout, "timeout", "", "Per-test timeout override (e.g. 30s)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "List tests without running")
	runCmd.Flags().BoolVar(&watch, "watch", false, "Re-run tests when files change (Ctrl+C to exit)")
	runCmd.Flags().StringVar(&envName, "env", "", "Load environment-specific config (e.g. staging loads staging.smoke.yaml)")
	runCmd.Flags().BoolVar(&monorepoMode, "monorepo", false, "Auto-discover .smoke.yaml in subdirectories")
}

func runSmoke(cmd *cobra.Command, args []string) error {
	cfg, err := schema.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Load environment-specific overrides
	if envName != "" {
		configDir := filepath.Dir(configFile)
		envFile := filepath.Join(configDir, envName+".smoke.yaml")
		cfg, err = schema.MergeEnv(cfg, envFile)
		if err != nil {
			return fmt.Errorf("loading env %q: %w", envName, err)
		}
	}

	if err := schema.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configDir := filepath.Dir(configFile)
	if !filepath.IsAbs(configDir) {
		cwd, _ := os.Getwd()
		configDir = filepath.Join(cwd, configDir)
	}

	// Check monorepo mode
	if monorepoMode || cfg.Settings.Monorepo {
		// Create reporter early for monorepo mode
		var rep reporter.Reporter
		switch format {
		case "json":
			rep = reporter.NewJSON(os.Stdout)
		case "junit":
			rep = reporter.NewJUnit(os.Stdout)
		case "tap":
			rep = reporter.NewTAP(os.Stdout)
		case "prometheus":
			rep = reporter.NewPrometheus(os.Stdout)
		default:
			rep = reporter.NewTerminal(os.Stdout)
		}

		configs, err := monorepo.Discover(configDir, cfg.Settings.MonorepoExclude)
		if err != nil {
			return fmt.Errorf("discovering sub-configs: %w", err)
		}
		if len(configs) == 0 {
			return fmt.Errorf("no smoke configs found in %s", configDir)
		}
		r := &runner.Runner{Config: cfg, Reporter: rep, ConfigDir: configDir}

		// Parse timeout
		var timeoutDur time.Duration
		if timeout != "" {
			timeoutDur, err = time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout %q: %w", timeout, err)
			}
		}

		if !watch {
			result, err := r.RunMonorepo(runner.RunOptions{
				Tags:        tags,
				ExcludeTags: excludeTags,
				FailFast:    failFast,
				DryRun:      dryRun,
				Timeout:     timeoutDur,
			}, configs)
			if err != nil {
				return err
			}
			if result.Failed > 0 {
				os.Exit(1)
			}
			return nil
		}

		return runWatch(configDir, configFile, func() error {
			_, err := r.RunMonorepo(runner.RunOptions{
				Tags:        tags,
				ExcludeTags: excludeTags,
				FailFast:    failFast,
				DryRun:      dryRun,
				Timeout:     timeoutDur,
			}, configs)
			if err != nil {
				return err
			}
			return nil
		})
	}

	// Create reporter
	var rep reporter.Reporter
	switch format {
	case "json":
		rep = reporter.NewJSON(os.Stdout)
	case "junit":
		rep = reporter.NewJUnit(os.Stdout)
	case "tap":
		rep = reporter.NewTAP(os.Stdout)
	case "prometheus":
		rep = reporter.NewPrometheus(os.Stdout)
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

	runOnce := func() error {
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
		// In watch mode, don't exit process on failure — print marker and loop
		if !watch && result.Failed > 0 {
			os.Exit(1)
		}
		return nil
	}

	if !watch {
		return runOnce()
	}

	return runWatch(configDir, configFile, runOnce)
}

func runWatch(configDir, configFile string, runOnce func() error) error {
	// Run once immediately
	if err := runOnce(); err != nil {
		fmt.Fprintf(os.Stderr, "initial run error: %v\n", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}
	defer w.Close()

	if err := w.Add(configDir); err != nil {
		return fmt.Errorf("watching %s: %w", configDir, err)
	}

	fmt.Fprintln(os.Stderr, "👀 watching", configDir, "(Ctrl+C to exit)")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Debounce: coalesce rapid events into a single re-run
	debounce := 500 * time.Millisecond
	var timer *time.Timer

	trigger := func() {
		fmt.Fprintln(os.Stderr, "\n🔁 change detected, re-running…")
		if err := runOnce(); err != nil {
			fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		}
	}

	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			// Ignore chmod-only events
			if !isRelevantEvent(ev.Op) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounce, trigger)
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "\n✋ stopping watcher")
			return nil
		}
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/detector"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate .smoke.yaml for this project",
	Long:  "Auto-detect project type and generate a .smoke.yaml configuration",
	RunE:  runInit,
}

var (
	forceOverwrite bool
	fromRunning    string
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite existing .smoke.yaml")
	initCmd.Flags().StringVar(&fromRunning, "from-running", "", "Generate config by inspecting a running Docker container")
}

func runInit(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(".smoke.yaml"); err == nil && !forceOverwrite {
		return fmt.Errorf(".smoke.yaml already exists (use --force to overwrite)")
	}

	var cfg *schema.SmokeConfig

	if fromRunning != "" {
		// Inspect running container
		fmt.Printf("Inspecting container: %s\n", fromRunning)
		var err error
		cfg, err = detector.InspectContainer(fromRunning)
		if err != nil {
			return fmt.Errorf("inspecting container: %w", err)
		}
		fmt.Printf("Found: %d ports, %d processes\n", len(cfg.Tests), countProcessTests(cfg))
	} else {
		// Auto-detect from filesystem
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		types := detector.Detect(cwd)
		if len(types) == 0 {
			fmt.Println("No project type detected. Creating a minimal .smoke.yaml")
		} else {
			names := make([]string, len(types))
			for i, t := range types {
				names[i] = string(t)
			}
			fmt.Printf("Detected: %v\n", names)
		}

		cfg = detector.GenerateConfig(cwd, types)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(".smoke.yaml", data, 0644); err != nil {
		return fmt.Errorf("writing .smoke.yaml: %w", err)
	}

	fmt.Println("Created .smoke.yaml")
	return nil
}

func countProcessTests(cfg *schema.SmokeConfig) int {
	count := 0
	for _, t := range cfg.Tests {
		if t.Expect.PortListening != nil {
			count++
		}
	}
	return count
}

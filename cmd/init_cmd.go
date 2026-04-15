package cmd

import (
	"fmt"
	"os"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/detector"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate .smoke.yaml for this project",
	Long:  "Auto-detect project type and generate a .smoke.yaml configuration",
	RunE:  runInit,
}

var forceOverwrite bool

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite existing .smoke.yaml")
}

func runInit(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(".smoke.yaml"); err == nil && !forceOverwrite {
		return fmt.Errorf(".smoke.yaml already exists (use --force to overwrite)")
	}

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

	cfg := detector.GenerateConfig(cwd, types)

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

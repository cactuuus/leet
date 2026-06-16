package cmd

import (
	"fmt"
	"strconv"

	"github.com/cactuuus/leet/internal/scaffold"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [number]",
	Short: "Open a problem folder (or the problems directory if no number is provided), in your editor.",
	Args:  cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := initPackages()
		if err != nil {
			return err
		}

		// no argument: open the root problems directory
		if len(args) == 0 {
			return cfg.OpenInEditor(cfg.ProblemsPath())
		}

		// argument given: resolve and open that specific problem's folder
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid problem number: %s", args[0])
		}
		dir, err := scaffold.GetProblemDirByNumber(number)
		if err != nil {
			return fmt.Errorf("%w — run 'leet load %d' first", err, number)
		}
		return cfg.OpenInEditor(dir)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}

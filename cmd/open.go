package cmd

import (
	"fmt"
	"strconv"

	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/scaffold"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [number|daily]",
	Short: "Open a problem folder or the problems directory, in your editor.",
	Args:  cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// load config, client and scaffolder
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		scaffolder, err := scaffold.NewScaffolder(cfg.ProblemsDir)
		if err != nil {
			return fmt.Errorf("failed to create scaffolder: %w", err)
		}

		// determine the directory to open, then open it in the editor
		var dir string
		switch {
			case len(args) == 0: // no argument: open the root problems directory
				dir = cfg.ProblemsDir
			case args[0] == "daily": // if the argument is "daily", find the daily problem directory and open it
				// create a new leetcode client to fetch the daily problem
				c, err := leetcode.NewClient()
				if err != nil {
					return fmt.Errorf("failed to create leetcode client: %w", err)
				}
				// fetch the daily problem
				fmt.Print("Fetching daily problem... ")
				p, err := c.FetchDailyProblem()
				if err != nil {
					return fmt.Errorf("failed to fetch daily problem: %w", err)
				}
				fmt.Print("✓\n")
				// update dir to the daily problem's directory
				dir, err = scaffolder.GetProblemDirByNumber(p.Number)
				if err != nil {
					return fmt.Errorf("%w — run 'leet load daily --open' instead", err)
				}
			default: // else, the argument should be a number, in which case we just open it
				number, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid problem number: %s", args[0])
				}
				dir, err = scaffolder.GetProblemDirByNumber(number)
				if err != nil {
					return fmt.Errorf("%w — run 'leet load %d --open' instead", err, number)
				}
		}

		fmt.Print("Opening in editor... ")
		err = openInEditor(cfg, dir)
		if err != nil {
			return fmt.Errorf("failed to open directory in editor: %w", err)
		}
		fmt.Print("✓\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}

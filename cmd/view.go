package cmd

import (
	"fmt"
	"strconv"

	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <number|daily>",
	Short: "Display a problem preview in the terminal.",
	Long: `Display a problem preview in the terminal.
Pass a problem number to preview a specific problem, or 'daily' for today's challenge.`,
	Args: cobra.ExactArgs(1),
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return openProblem(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func openProblem(_ *cobra.Command, args []string) error {
	// load config and initialize leetcode package
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if err := leetcode.Init(cfg); err != nil {
		return err
	}

	var problem leetcode.Problem
	if args[0] == "daily" {
		problem, err = leetcode.FetchDailyProblem()
		if err != nil {
			return err
		}
	} else {
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid problem number: %s — use a number or 'daily'", args[0])
		}
		problem, err = leetcode.FetchProblem(number)
		if err != nil {
			return err
		}
	}

	printProblem(problem)
	return nil
}

// printProblem displays a formatted problem preview in the terminal.
func printProblem(problem leetcode.Problem) {
	fmt.Printf("%d. %s\t[%s]", problem.Number, problem.Name, problem.Difficulty)
	fmt.Println()
	fmt.Println()
	fmt.Println(problem.Link())
	fmt.Println()
	fmt.Println(problem.Summary(500))
	fmt.Println()
}

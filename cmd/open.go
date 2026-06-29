package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewOpenCmd(ctx AppContext) *cobra.Command {
	var openCmd = &cobra.Command{
		Use:   			"open [number|daily]",
		Short: 			"Open a problem folder, or the problems directory, in your editor.",
		Args:  			cobra.MaximumNArgs(1),
		SilenceUsage:  	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			return openProblem(cmd, args, ctx)
		},
	}

	return openCmd
}

func openProblem(_ *cobra.Command, args []string, ctx AppContext) error {
	cfg := ctx.Config()

	var dirToOpen string
	// if no argument is provided, open the root problems directory
	if len(args) == 0 {
		dirToOpen = cfg.ProblemsDir
	} else {
		c, s := ctx.Client(), ctx.Scaffolder()
		// get the problem preview based on the provided argument (number or "daily")
		preview, err := fetchPreviewByIdentifier(c, args[0])
		if err != nil {
			return err
		}
		dirToOpen = s.GetProblemDir(preview)

		// check if the problem folder exists
		_, err = os.Stat(dirToOpen)
		if os.IsNotExist(err) {
			return fmt.Errorf("problem not loaded yet, use 'leet load %s --open' instead", args[0])
		} else if err != nil {
			return fmt.Errorf("failed to check if problem folder exists: %w", err)
		}
	}

	fmt.Print("Opening in editor... ")
	if err := openInEditor(cfg.Editor, dirToOpen); err != nil {
		return fmt.Errorf("failed to open directory in editor: %w", err)
	}
	fmt.Print("✓\n")
	return nil
}

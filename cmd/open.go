package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewOpenCmd(ctx AppContext) *cobra.Command {
	var openCmd = &cobra.Command{
		Use:   	"open [number|daily]",
		Short:	"Open a problem folder, or the problems directory, in your editor.",
		Args:	cobra.MaximumNArgs(1),
		RunE: 	func(cmd *cobra.Command, args []string) error {
			return openProblem(cmd, args, ctx)
		},
	}

	return openCmd
}

func openProblem(_ *cobra.Command, args []string, ctx AppContext) error {
	cfg := ctx.Config()
	dirToOpen := cfg.ProblemsDir
	// parse the optional argument to determine which directory to open
	if len(args) > 0 {
		c, err := ctx.Client()
		if err != nil {
			return err
		}
		s, err := ctx.Scaffolder()
		if err != nil {
			return err
		}
		preview, err := fetchPreviewByIdentifier(c, args[0])
		if err != nil {
			return err
		}
		dirToOpen = s.GetProblemDir(preview)

		// check if the problem folder exists
		_, err = os.Stat(dirToOpen)
		if os.IsNotExist(err) {
			return fmt.Errorf("Problem not loaded yet, use 'leet load %s --open' instead", args[0])
		} else if err != nil {
			return fmt.Errorf("Failed to check if problem folder exists:\n%w", err)
		}
	}

	return openInEditor(cfg, dirToOpen)
}

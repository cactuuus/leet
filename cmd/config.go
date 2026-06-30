package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewConfigCmd(ctx AppContext) *cobra.Command {
	// maing config command
	configCmd := &cobra.Command{
		Use:   	"config",
		Short: 	"View or edit leet's configuration.",
		Long: 	fmt.Sprintf(
				"View or edit leet's configuration.\n" +
				"The config file lives at %s and is meant to be edited directly. " +
				"Run 'leet config edit' to open it in your editor, " +
				"or 'leet config reset' to restore it to default values.",
				ctx.Config().Path),
	}

	// register subcommands
	configCmd.AddCommand(newConfigShowCmd(ctx))
	configCmd.AddCommand(newConfigEditCmd(ctx))
	configCmd.AddCommand(newConfigResetCmd(ctx))

	return configCmd
}

func newConfigShowCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   	"show",
		Short: 	"Print current configuration.",
		Long: 	"Print current configuration.\nAny sensitive information, such as " +
				"your LeetCode credentials, will not be printed for security reasons.",
		RunE: 	func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(ctx.Config().String())
			return nil
		},
	}
}

func newConfigEditCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:	"edit",
		Short: 	"Open the config file in your editor, for manual editing.",
		Long:	fmt.Sprintf(
				"Open the config file in your editor, for manual editing.\n" +
				"You can otherwise access it directly at %s.",
				ctx.Config().Path),
		RunE: 	func(cmd *cobra.Command, args []string) error {
			cfg := ctx.Config()
			if err := openInEditor(cfg.Editor, cfg.Path); err != nil {
				return fmt.Errorf(
					"Failed to open config file in editor. Open it manually at %s, " +
					"or run 'leet config reset' to restore it to default values:\n%w",
					cfg.Path, err)
			}
			return nil
		},
	}
}

func newConfigResetCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:	"reset",
		Short: 	"Reset the config file to default values.",
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			confirmed, err := promptYesNo(
				"Are you sure you want to reset the config file to default values? " +
				"This action cannot be undone.",
			)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Aborted.")
				return nil
			}
			printActionStart("Resetting config file to default values")
			if err := ctx.Config().Reset(); err != nil {
				return err
			}
			printActionSuccess()
			return nil
		},
	}


}

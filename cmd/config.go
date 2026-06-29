package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewConfigCmd(ctx AppContext) *cobra.Command {
	// maing config command
	configCmd := &cobra.Command{
		Use:   			"config",
		Short: 			"View or edit leet's configuration.",
		Long: 			fmt.Sprintf("View or edit leet's configuration.\nThe config file lives at %s and is meant to be edited directly. Run 'leet config edit' to open it in your editor, or 'leet config reset' to restore it to default values.", ctx.Config().Path),
		SilenceUsage: 	true,
	}

	// subcommands

	configShowCmd := &cobra.Command{
		Use:   			"show",
		Short: 			"Print current configuration.",
		Long: 			"Print current configuration.\nAny sensitive information, such as your LeetCode credentials, will not be printed for security reasons.",
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			fmt.Println(ctx.Config().String())
			return nil
		},
	}

	configEditCmd := &cobra.Command{
		Use:   			"edit",
		Short: 			"Open the config file in your editor, for manual editing.",
		Long:			fmt.Sprintf("Open the config file in your editor, for manual editing.\nThis uses the editor_cmd specified in your config file. You can otherwise access it directly at %s.", ctx.Config().Path),
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			fmt.Print("Opening config file in editor... ")
			cfg := ctx.Config()
			if err := openInEditor(cfg.Editor, cfg.Path); err != nil {
				return fmt.Errorf(
					"failed to open config file in editor: %w\n" +
					"Try opening it manually at %s, or run 'leet config reset' to restore it to default values.",
					err, cfg.Path)
			}
			fmt.Print("✓\n")
			return nil
		},
	}

	configResetCmd := &cobra.Command{
		Use:   			"reset",
		Short: 			"Reset the config file to default values.",
		Long: 			"Reset the config file to default values.\nThis simply deletes the config file and creates a new one with default values, hence it can be used to fix a corrupted config file.",
		SilenceUsage: 	true,
		RunE: 			func(cmd *cobra.Command, args []string) error {
			confirmed, err := promptYesNo(
				"Are you sure you want to reset the config file to default values?" +
				" This action cannot be undone.",
			)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Aborted.")
				return nil
			}
			fmt.Print("Resetting config file to default values... ")
			if err := ctx.Config().Reset(); err != nil {
				return err
			}
			fmt.Print("✓\n")
			return nil
		},
	}

	// register subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configResetCmd)

	return configCmd
}

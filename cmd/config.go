package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cactuuus/leet/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or edit leet's configuration.",
	Long: `View or edit leet's configuration.
The config file lives at ~/.config/leet/config.toml and is meant to be edited directly — there
are no individual 'set' commands. It's commented with what each setting does; 'leet config edit'
opens it in your editor.`,
	SilenceUsage: true,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println(cfg)
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in your editor, for manual editing.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get the path to the config file
		path, err := config.Path()
		if err != nil {
			return err
		}
		// load the config - if it fails, it prints a warning but still attempts to open the config file in the editor, so the user can fix it manually.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			fmt.Printf("Attempting to open the config file anyway, if this fails, you can manually edit it at '%s', or run 'leet config reset' to restore it to default values.\n", path)
		}
		fmt.Print("Opening in editor... ")
		if err := openInEditor(cfg, path); err != nil {
			return fmt.Errorf("failed to open config file in editor: %w", err)
		}
		fmt.Print("✓\n")
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the config file to default values. This simply deletes the config file and creates a new one with default values, hence it can be used to fix a corrupted config file.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Are you sure you want to reset the config file to default values? This action cannot be undone. [y/n]")
		scanner := bufio.NewScanner(os.Stdin)
		loop := true
		for loop {
			fmt.Print("> ")
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				return fmt.Errorf("no input received, aborting")
			}
			switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
			case "y", "yes":
				// proceed with reset
				loop = false
			case "n", "no":
				loop = false
				return nil
			default:
				fmt.Println("Please enter 'y' or 'n'.")
			}
		}

		fmt.Print("Resetting config file to default values... ")
		if err := config.Reset(); err != nil {
			return err
		}
		fmt.Print("✓\n")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}

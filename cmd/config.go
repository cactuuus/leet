package cmd

import (
	"fmt"

	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/language"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage leet configuration.",
	Long:  `View and update leet configuration settings.`,
	SilenceUsage:  true,
}

// config show, print current config values
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration.",
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		fmt.Printf("Problems directory : %s\n", cfg.Paths.Problems)
		fmt.Printf("Cache path         : %s\n", cfg.Paths.Cache)
		fmt.Printf("Default languages  : %v\n", cfg.Languages.Preferred)
		return nil
	},
}

// config set-languages, update the default languages
var configSetLanguagesCmd = &cobra.Command{
	Use:   "set-languages <languages...>",
	Short: "Set the default languages to scaffold.",
	Long: `Set the default languages used when no languages are specified in 'leet load'.
Accepts language slugs or names, e.g: golang, python3, typescript, Go, C++`,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("at least one language is required")
		}

		// validate all languages before making any changes
		var slugs []string
		for _, arg := range args {
			l, ok := language.Get(arg)
			if !ok {
				return fmt.Errorf("unknown language: %q — run 'leet languages' to see supported languages", arg)
			}
			slugs = append(slugs, l.Slug)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.Languages.Preferred = slugs
		if err := config.UpdateConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Default languages updated: %v\n", slugs)
		return nil
	},
}

// config set-problems-dir, update the problems directory
var configSetProblemsDirCmd = &cobra.Command{
	Use:   "set-problems-dir <path>",
	Short: "Set the directory where problems are scaffolded.",
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("path is required")
		}
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.Paths.Problems = args[0]
		if err := config.UpdateConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Problems directory updated: %s\n", args[0])
		return nil
	},
}

var configSetEditorCmd = &cobra.Command{
	Use:   "set-editor-cmd <command>",
	Short: "Set the command used to open problem folders (e.g. 'code', 'subl', 'nvim').",
	Args:  cobra.ExactArgs(1),
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.Editor.Command = args[0]
		if err := config.UpdateConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Editor command set to: %s\n", args[0])
		return nil
	},
}

func init() {
	// register subcommands on configCmd
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetLanguagesCmd)
	configCmd.AddCommand(configSetProblemsDirCmd)
	configCmd.AddCommand(configSetEditorCmd)
	// register configCmd with root
	rootCmd.AddCommand(configCmd)
}

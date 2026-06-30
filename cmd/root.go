package cmd

import (
	"os"

	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/scaffold"
	"github.com/spf13/cobra"
)

// AppContext defines the interface for the main application, providing access to its core components.
type AppContext interface {
	Config() 		*config.Manager
	Client() 		*leetcode.Client
	Scaffolder() 	*scaffold.Scaffolder
}

// Execute initializes each command and runs the root command of the application.
func Execute(ctx AppContext) {
	rootCmd := &cobra.Command{
		Use:   "leet",
		Short: "A simple utility to streamline leetcoding locally.",
		Long:  `Leet is a simple application that allows to more effectively leetcode locally.`,
	}

	// we register the subcommands here, passing the app to each of them
	rootCmd.AddCommand(NewLoadCmd(ctx))
	rootCmd.AddCommand(NewOpenCmd(ctx))
	rootCmd.AddCommand(NewLanguagesCmd(ctx))
	rootCmd.AddCommand(NewConfigCmd(ctx))
	rootCmd.AddCommand(NewTestCmd(ctx))
	rootCmd.AddCommand(NewTemplateCmd(ctx))
	rootCmd.AddCommand(NewSubmitCmd(ctx))
	rootCmd.AddCommand(NewCacheCmd(ctx))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

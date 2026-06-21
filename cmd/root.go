package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "leet",
	Short: "A simple utility to streamline leetcoding locally.",
	Long: `Leet is a simple application that allows to more effectively leetcode locally.
Given a problem number, it creates a directory containing one file per specified language,
all initialized with the relevant code snippet. This removes the clunky process of switiching
between languages in the online interface, and copy/pasting into your own editor.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.leet.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

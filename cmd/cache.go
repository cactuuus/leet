package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCacheCmd(ctx AppContext) *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:          "cache",
		Short:        "Manage the local cache of LeetCode problems.",
	}

	cacheCmd.AddCommand(newCacheClearCmd(ctx))
	cacheCmd.AddCommand(newCacheSummaryCmd(ctx))
	cacheCmd.AddCommand(newCacheOpenCmd(ctx))

	return cacheCmd
}

func newCacheClearCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   	"clear",
		Short: 	"Clear the local cache.",
		Args:  	cobra.NoArgs,
		RunE: 	func(cmd *cobra.Command, args []string) error {
			c, err := ctx.Client()
			if err != nil {
				return err
			}
			printActionStart("Clearing cache")
			if err := c.ClearCache(); err != nil {
				return fmt.Errorf("Failed to clear cache:\n%w", err)
			}
			printActionSuccess()
			return nil
		},
	}
}

func newCacheSummaryCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   	"summary",
		Short: 	"Show a summary of the local cache.",
		Args:  	cobra.NoArgs,
		RunE: 	func(cmd *cobra.Command, args []string) error {
			c, err := ctx.Client()
			if err != nil {
				return err
			}
			summary, err := c.CacheSummary()
			if err != nil {
				return fmt.Errorf(
					"Failed to get cache summary. " +
					"Run 'leet cache clear' if the issue persists:\n%w",
					err)
			}
			fmt.Println()
			fmt.Println(summary)
			return nil
		},
	}
}

func newCacheOpenCmd(ctx AppContext) *cobra.Command {
	path := ctx.Config().CachePath

	return &cobra.Command{
		Use:   	"open",
		Short: 	"Open the cache directory in your editor.",
		Long:  	fmt.Sprintf(
				"Open the cache directory in your editor.\n" +
				"You can otherwise access it directly at %s.",
				path),
		Args:  	cobra.NoArgs,
		RunE: 	func(cmd *cobra.Command, args []string) error {
			if err := openInEditor(ctx.Config(), path); err != nil {
				return fmt.Errorf(
					"Failed to open cache directory in editor. Open it manually at %s, " +
					"or run 'leet cache clear' to reset the cache:\n%w",
					path, err)
			}
			return nil
		},
	}
}

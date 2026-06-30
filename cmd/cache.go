package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCacheCmd(ctx AppContext) *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:          "cache",
		Short:        "Manage the local cache of LeetCode problems.",
		SilenceUsage: true,
	}

	cacheClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the local cache.",
		Long:  "Clear the local cache of LeetCode problems. This will remove all cached problem data and force leet to fetch fresh data from LeetCode on the next operation.",
		Args:  cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Client().ClearCache()
		},
	}

	cacheSummaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "Show a summary of the local cache.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := ctx.Client().CacheSummary()
			if err != nil {
				return fmt.Errorf("failed to get cache summary, try running 'leet cache clear' if the issue persists: %w", err)
			}
			fmt.Println()
			fmt.Println(summary)
			return nil
		},
	}

	cacheOpenCmd := &cobra.Command{
		Use:   "open",
		Short: "Open the cache directory in your editor.",
		Long:  fmt.Sprintf("Open the cache directory in your editor.\nYou can otherwise access it directly at %s.", ctx.Client().CachePath()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Opening cache directory in editor... ")
			if err := openInEditor(ctx.Config().Editor, ctx.Client().CachePath()); err != nil {
				return fmt.Errorf(
					"failed to open cache directory in editor: %w\n" +
						"Try opening it manually at %s, or run 'leet cache clear' to reset the cache.",
					err, ctx.Client().CachePath())
			}
			fmt.Print("✓\n")
			return nil
		},
	}

	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheSummaryCmd)
	cacheCmd.AddCommand(cacheOpenCmd)

	return cacheCmd
}

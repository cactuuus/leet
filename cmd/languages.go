package cmd

import (
	"fmt"
	"sort"

	"github.com/cactuuus/leet/internal/language"
	"github.com/spf13/cobra"
)

func NewLanguagesCmd(ctx AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   	"languages",
		Short:	"List all supported languages.",
		RunE: 	func(cmd *cobra.Command, args []string) error {
			langs := language.All()
			// sort alphabetically by name for consistent output
			sort.Slice(langs, func(i, j int) bool {
				return langs[i].Name < langs[j].Name
			})
			fmt.Printf("%-15s %s\n", "NAME", "SLUG")
			fmt.Printf("%-15s %s\n", "----", "----")
			for _, l := range langs {
				fmt.Printf("%-15s %s\n", l.Name, l.Slug)
			}
			return nil
		},
	}
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/scaffold"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load <number|daily>",
	Short: "Load a problem from LeetCode and scaffold it locally.",
	Long: `Load a problem from LeetCode and scaffold it locally.
Given a problem number or 'daily', creates a directory containing the problem
description and one file per language, initialized with the relevant starter code.`,
	SilenceUsage:  true,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return loadProblem(cmd, args)
	},
}

func init() {
	loadCmd.Flags().StringSliceP("langs", "l", nil, "Languages to scaffold (comma-separated). Defaults to configured preferred languages.")
	loadCmd.Flags().BoolP("desc-only", "d", false, "Only create the problem description, skip code files.")
	loadCmd.Flags().BoolP("force", "f", false, "Overwrite existing files without prompting.")
	loadCmd.Flags().BoolP("open", "o", false, "Open the problem folder after scaffolding.")
	rootCmd.AddCommand(loadCmd)
}

// loadProblem is the entry point for 'leet load'.
func loadProblem(cmd *cobra.Command, args []string) error {
	// initialize config, client, and scaffolder
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	client, err := leetcode.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create leetcode client: %w", err)
	}
	scaffolder, err := scaffold.NewScaffolder(cfg.ProblemsDir)
	if err != nil {
		return fmt.Errorf("failed to create scaffolder: %w", err)
	}

	// fetch the problem from LeetCode
	fmt.Print("Fetching problem... ")
	problem, err := fetchByIdentifier(client, args[0])
	if err != nil {
		return err
	}
	fmt.Print("✓\n")

	// determine which languages to scaffold
	langs, err := resolveLanguages(cmd, cfg)
	if err != nil {
		return err
	}

	// if --force is not set and there are languages to scaffold, check for conflicts
	force, _ := cmd.Flags().GetBool("force")
	if !force && len(langs) > 0 {
		langs, err = resolveConflicts(scaffolder, problem, langs)
		if err != nil {
			return err
		}
	}

	// create description
	fmt.Print("Creating description... ")
	if err := scaffolder.CreateDescription(problem); err != nil {
		return fmt.Errorf("failed to create description: %w", err)
	}
	fmt.Print("✓\n")

	// create code files for each language
	for _, l := range langs {
		fmt.Printf("Creating %s snippet... ", l.Name)
		if err := scaffolder.CreateSnippet(problem, l); err != nil {
			return fmt.Errorf("failed to create snippet for %s: %w", l.Name, err)
		}
		fmt.Print("✓\n")
	}

	fmt.Printf("Scaffolded problem %d (%s)\n", problem.Number, problem.Name)

	// open in editor if requested
	if open, _ := cmd.Flags().GetBool("open"); open {
		fmt.Print("Opening in editor... ")
		if err := openInEditor(cfg, scaffolder.GetProblemDir(problem)); err != nil {
			return fmt.Errorf("failed to open problem folder: %w", err)
		}
		fmt.Print("✓\n")
	}
	return nil
}

// fetchByIdentifier fetches a problem given either "daily" or a numeric string.
func fetchByIdentifier(c *leetcode.Client, id string) (leetcode.Problem, error) {
	if id == "daily" {
		return c.FetchDailyProblem()
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return leetcode.Problem{}, fmt.Errorf("invalid problem identifier: %q — use a problem number or 'daily'", id)
	}
	return c.FetchProblem(num)
}

// resolveLanguages determines which languages to scaffold based on flags and config.
// Precedence, most specific first:
//  1. --desc-only           -> no languages (description only)
//  2. --langs (provided)    -> exactly what was passed, even if empty
//  3. neither flag provided -> fall back to config.Languages.Preferred
func resolveLanguages(cmd *cobra.Command, cfg config.Config) ([]language.Language, error) {
	descOnly, _ := cmd.Flags().GetBool("desc-only")
	if descOnly {
		return nil, nil
	}

	if cmd.Flags().Changed("langs") {
		raw, _ := cmd.Flags().GetStringSlice("langs")
		return parseLanguageArgs(raw)
	}

	if len(cfg.PreferredLanguages) == 0 {
		return nil, fmt.Errorf("no languages specified and no defaults configured — use --langs, set defaults in your config file (leet config edit)")
	}
	return parseLanguageArgs(cfg.PreferredLanguages)
}

// parseLanguageArgs resolves a list of language slugs/names into Language structs, deduplicating
// and validating each one. Returns an empty slice for an empty input.
func parseLanguageArgs(raw []string) ([]language.Language, error) {
	seen := make(map[string]struct{})
	langs := make([]language.Language, 0, len(raw))

	for _, id := range raw {
		l, ok := language.Get(id)
		if !ok {
			return nil, fmt.Errorf("unknown language: %q — run 'leet languages' to see supported languages", id)
		}
		if _, dup := seen[l.Slug]; dup {
			continue
		}
		seen[l.Slug] = struct{}{}
		langs = append(langs, l)
	}
	return langs, nil
}

// resolveConflicts checks for existing files among langs and, if any are found, prompts the user to
// overwrite, skip the conflicting ones, or abort.
func resolveConflicts(s *scaffold.Scaffolder, p leetcode.Problem, langs []language.Language) ([]language.Language, error) {
	conflicts, err := findConflicts(s, p, langs)
	if err != nil {
		return nil, err
	}
	if len(conflicts) == 0 {
		return langs, nil
	}

	fmt.Println("The following files already exist:")
	for _, l := range conflicts {
		fmt.Printf("  - %s\n", s.GetSnippetFilename(p, l))
	}
	fmt.Println("[y: overwrite, n: skip, a: abort]")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("failed to read input: %w", err)
			}
			return nil, fmt.Errorf("no input received, aborting")
		}

		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "y", "yes":
			return langs, nil
		case "n", "no":
			return removeConflicting(langs, conflicts), nil
		case "a", "abort":
			return nil, fmt.Errorf("aborted by user")
		default:
			fmt.Println("Please enter 'y', 'n', or 'a'")
		}
	}
}

// findConflicts returns the languages that already have a snippet file on disk for this problem.
func findConflicts(s *scaffold.Scaffolder, p leetcode.Problem, langs []language.Language) ([]language.Language, error) {
	var conflicts []language.Language
	for _, l := range langs {
		exists, err := s.SnippetExists(p, l)
		if err != nil {
			return nil, err
		}
		if exists {
			conflicts = append(conflicts, l)
		}
	}
	return conflicts, nil
}

// removeConflicting returns langs with any entries present in conflicts removed.
func removeConflicting(langs, conflicts []language.Language) []language.Language {
	skip := make(map[string]struct{}, len(conflicts))
	for _, l := range conflicts {
		skip[l.Slug] = struct{}{}
	}

	filtered := make([]language.Language, 0, len(langs))
	for _, l := range langs {
		if _, isConflict := skip[l.Slug]; !isConflict {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

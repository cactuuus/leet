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

// loadProblem is the entry point for 'leet load'. Each step is broken into its
// own function so the overall flow reads top to bottom without nested logic.
func loadProblem(cmd *cobra.Command, args []string) error {
	cfg, err := initPackages()
	if err != nil {
		return err
	}

	problem, err := fetchByIdentifier(args[0])
	if err != nil {
		return err
	}

	langs, err := resolveLanguages(cmd, cfg)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && len(langs) > 0 {
		langs, err = resolveConflicts(problem, langs)
		if err != nil {
			return err
		}
	}

	if err := scaffold.ScaffoldProblem(problem, langs); err != nil {
		return err
	}

	printLoadSummary(problem, langs)

	// open in editor if requested
	if open, _ := cmd.Flags().GetBool("open"); open {
		if err := cfg.OpenInEditor(scaffold.GetProblemDir(problem)); err != nil {
			return fmt.Errorf("failed to open problem folder: %w", err)
		}
	}
	return nil
}

// initPackages loads the config file and initializes the packages that depend
// on it (leetcode, scaffold). Returns the loaded config for callers that need it.
// TODO: remove init, instead use constructors
func initPackages() (config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.Config{}, err
	}
	if err := scaffold.Init(cfg); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

// fetchByIdentifier fetches a problem given either "daily" or a numeric string.
func fetchByIdentifier(identifier string) (leetcode.Problem, error) {
	c, err := leetcode.NewClient()
	if err != nil {
		return leetcode.Problem{}, fmt.Errorf("failed to create leetcode client: %w", err)
	}

	if identifier == "daily" {
		return c.FetchDailyProblem()
	}
	number, err := strconv.Atoi(identifier)
	if err != nil {
		return leetcode.Problem{}, fmt.Errorf("invalid problem identifier: %q — use a problem number or 'daily'", identifier)
	}
	return c.FetchProblem(number)
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

	if len(cfg.Languages.Preferred) == 0 {
		return nil, fmt.Errorf("no languages specified and no defaults configured — use --langs, set defaults with 'leet config set-languages', or use --desc-only")
	}
	return parseLanguageArgs(cfg.Languages.Preferred)
}

// parseLanguageArgs resolves a list of language slugs/names into Language structs,
// deduplicating and validating each one. Returns an empty slice for an empty input.
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

// resolveConflicts checks for existing files among langs and, if any are found,
// prompts the user to overwrite, skip the conflicting ones, or abort.
func resolveConflicts(problem leetcode.Problem, langs []language.Language) ([]language.Language, error) {
	conflicts, err := scaffold.CheckConflicts(problem, langs)
	if err != nil {
		return nil, err
	}
	if len(conflicts) == 0 {
		return langs, nil
	}

	fmt.Println("The following files already exist:")
	for _, l := range conflicts {
		fmt.Printf("  - %s\n", scaffold.GetFilename(problem, l))
	}
	fmt.Println("[y: overwrite, n: skip them, a: abort]")

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

// printLoadSummary prints a short confirmation of what was scaffolded.
func printLoadSummary(problem leetcode.Problem, langs []language.Language) {
	fmt.Printf("Scaffolded problem %d (%s)\n", problem.Number, problem.Name)
	if len(langs) == 0 {
		fmt.Println("Description only — no code files created.")
		return
	}
	names := make([]string, len(langs))
	for i, l := range langs {
		names[i] = l.Name
	}
	fmt.Printf("Languages: %s\n", strings.Join(names, ", "))
}

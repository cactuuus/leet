package cmd

import (
	"fmt"
	"strconv"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/problem"
	"github.com/spf13/cobra"
)

func NewLoadCmd(ctx AppContext) *cobra.Command {
	var loadCmd = &cobra.Command{
		Use:   			"load <number|daily>",
		Short: 			"Load a problem from LeetCode and scaffold it locally.",
		Long: 			"Load a problem from LeetCode and scaffold it locally.\nGiven a problem number or 'daily', creates a directory containing the problem description, example test cases, and coding templates for the specified languages.",
		SilenceUsage: 	true,
		Args: 			cobra.ExactArgs(1),
		RunE: 			func(cmd *cobra.Command, args []string) error {
			return loadProblem(cmd, args, ctx)
		},
	}

	// define flags
	loadCmd.Flags().StringSliceP(
		"langs", "l", nil,
		"Languages to scaffold (comma-separated). Defaults to configured preferred languages.",
	)
	loadCmd.Flags().BoolP("no-desc", "d", false, "Skip creating the problem description.")
	loadCmd.Flags().BoolP("no-tests", "t", false, "Skip creating the testcases file.")
	loadCmd.Flags().BoolP("no-code", "c", false, "Skip creating code snippet files.")
	loadCmd.Flags().BoolP("force", "f", false, "Overwrite existing files without prompting.")
	loadCmd.Flags().BoolP("open", "o", false, "Open the problem folder after scaffolding.")

	return loadCmd
}


func loadProblem(cmd *cobra.Command, args []string, ctx AppContext) error {
	// parse flags
	noDesc, err := cmd.Flags().GetBool("no-desc")
	if err != nil {
		return fmt.Errorf("failed to parse --no-desc flag: %w", err)
	}
	noTests, err := cmd.Flags().GetBool("no-tests")
	if err != nil {
		return fmt.Errorf("failed to parse --no-tests flag: %w", err)
	}
	noCode, err := cmd.Flags().GetBool("no-code")
	if err != nil {
		return fmt.Errorf("failed to parse --no-code flag: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to parse --force flag: %w", err)
	}

	// check for contradictory flags
	if noCode && cmd.Flags().Changed("langs") {
		return fmt.Errorf("--no-code and --langs are contradictory")
	}

	// get these early, so that if they fail we exit before doing any work
	c, s, cfg := ctx.Client(), ctx.Scaffolder(), ctx.Config()

	// fetch the problem from LeetCode
	fmt.Print("Fetching problem... ")
	p, err := fetchByIdentifier(c, args[0])
	if err != nil {
		return err
	}
	fmt.Print("✓\n")

	// create description, unless skipped
	if !noDesc {
		fmt.Print("Creating description... ")
		if err := s.WriteDescription(p); err != nil {
			return fmt.Errorf("failed to create description: %w", err)
		}
		fmt.Print("✓\n")
	}

	// create testcases, unless skipped
	if !noTests {
		exists, err := s.TestcasesExists(p.Preview)
		if err != nil {
			return fmt.Errorf("failed to check testcases: %w", err)
		}

		write := !exists
		if exists && !force {
			write, err = promptYesNo("Testcases file already exists, overwrite?")
			if err != nil {
				return err
			}
		} else if exists && force {
			write = true
		}

		if write {
			fmt.Print("Creating testcases... ")
			if err := s.WriteTestcases(p.Preview, p.ExampleTestcases); err != nil {
				return fmt.Errorf("failed to create testcases: %w", err)
			}
			fmt.Print("✓\n")
		} else {
			fmt.Println("Skipping testcases. ✓")
		}
	}

	// create code snippets, unless skipped
	if !noCode {
		langs, err := resolveLanguages(cmd, cfg.PreferredLanguages)
		if err != nil {
			return err
		}

		for _, l := range langs {
			exists, err := s.SnippetExists(p.Preview, l)
			if err != nil {
				return err
			}

			write := !exists
			if exists && !force {
				write, err = promptYesNo(fmt.Sprintf("%s snippet already exists, overwrite?", l.Name))
				if err != nil {
					return err
				}
			} else if exists && force {
				write = true
			}

			if write {
				fmt.Printf("Creating %s snippet... ", l.Name)
				if err := s.WriteSnippet(p, l); err != nil {
					return fmt.Errorf("failed to create snippet for %s: %w", l.Name, err)
				}
				fmt.Print("✓\n")
			} else {
				fmt.Printf("Skipping %s snippet.\n", l.Name)
			}
		}
	}

	fmt.Printf("Scaffolded problem %d (%s)\n", p.Number, p.Title)

	// open in editor if requested
	if open, _ := cmd.Flags().GetBool("open"); open {
		fmt.Print("Opening in editor... ")
		if err := openInEditor(cfg.Editor, s.GetProblemDir(p.Preview)); err != nil {
			return fmt.Errorf("failed to open problem folder: %w", err)
		}
		fmt.Print("✓\n")
	}
	return nil
}

// fetchByIdentifier fetches a problem and its testcases given either "daily" or a numeric string.
func fetchByIdentifier(c *leetcode.Client, id string) (problem.Full, error) {
	if id == "daily" {
		return c.FetchDailyProblem()
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return problem.Full{}, fmt.Errorf("invalid problem identifier: %q — use a problem number or 'daily'", id)
	}
	return c.FetchProblem(num)
}


// resolveLanguages determines which languages to scaffold based on flags and config.
func resolveLanguages(cmd *cobra.Command, preferred []string) ([]language.Language, error) {
	if cmd.Flags().Changed("langs") {
		raw, _ := cmd.Flags().GetStringSlice("langs")
		return parseLanguageArgs(raw)
	}
	if len(preferred) == 0 {
		return nil, fmt.Errorf("no languages specified and no defaults configured — use --langs, or set defaults with 'leet config edit'")
	}
	return parseLanguageArgs(preferred)
}

// parseLanguageArgs resolves a list of language slugs/names into Language structs, deduplicating
// and validating each one.
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

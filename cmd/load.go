package cmd

import (
	"fmt"

	"github.com/cactuuus/leet/internal/language"
	"github.com/spf13/cobra"
)

type loadCmdBoolFlag int
const (
	noDescFlag loadCmdBoolFlag = iota
	noTestsFlag
	noCodeFlag
	forceFlag
	openFlag
)

var loadCmdBoolFlags = map[loadCmdBoolFlag]cmdBoolFlags {
	noDescFlag: {Long: "no-desc", Short: "d", Value: false, Desc: "Skip creating the problem description."},
	noTestsFlag: {Long: "no-tests", Short: "t", Value: false, Desc: "Skip creating the testcases file."},
	noCodeFlag: {Long: "no-code", Short: "c", Value: false, Desc: "Skip creating code snippet files."},
	forceFlag: {Long: "force", Short: "f", Value: false, Desc: "Overwrite existing files without prompting."},
	openFlag: {Long: "open", Short: "o", Value: false, Desc: "Open the problem folder after scaffolding."},
}

func NewLoadCmd(ctx AppContext) *cobra.Command {
	var loadCmd = &cobra.Command{
		Use:	"load <number|daily> [lang1 lang2 ...]",
		Short: 	"Load a problem from LeetCode and scaffold it locally.",
		Long: 	"Load a problem from LeetCode and scaffold it locally.\n" +
				"Given a problem number or 'daily', creates a directory containing the problem " +
				"description, example test cases, and coding templates for the specified languages.",
		Args:	cobra.MinimumNArgs(1),
		RunE:	func(cmd *cobra.Command, args []string) error {
			return loadProblem(cmd, args, ctx)
		},
	}

	// define flags
	for _, flag := range loadCmdBoolFlags {
		loadCmd.Flags().BoolP(flag.Long, flag.Short, flag.Value, flag.Desc)
	}
	return loadCmd
}

func loadProblem(cmd *cobra.Command, args []string, ctx AppContext) error {
	// parse flags
	flags := make(map[loadCmdBoolFlag]bool, len(loadCmdBoolFlags))
	for id, flag := range loadCmdBoolFlags {
		value, err := cmd.Flags().GetBool(flag.Long)
		if err != nil {
			return fmt.Errorf("Failed to parse --%s (-%s) flag:\n%w", flag.Long, flag.Short, err)
		}
		flags[id] = value
	}

	// check for contradictory arguments
	hasCustomLangs := len(args) > 1
	if flags[noCodeFlag] && hasCustomLangs {
		return fmt.Errorf(
			"The --no-code flag contradicts specifying languages." +
			"Please specify either --no-code and no languages, or a just list of languages.")
	}

	// get these early, so that if they fail we exit before doing any work
	cfg := ctx.Config()
	c, err := ctx.Client()
	if err != nil {
		return err
	}
	s, err := ctx.Scaffolder()
	if err != nil {
		return err
	}

	// fetch the problem from LeetCode
	p, err := fetchFullByIdentifier(c, args[0])
	if err != nil {
		return err
	}

	// create description, unless skipped
	if !flags[noDescFlag] {
		printActionStart("Creating description")
		if err := s.WriteDescription(p); err != nil {
			return fmt.Errorf("Failed to create description:\n%w", err)
		}
		printActionSuccess()
	}

	// create testcases, unless skipped
	if !flags[noTestsFlag] {
		exists, err := s.TestcasesExists(p.Preview)
		if err != nil {
			return fmt.Errorf("Failed to check testcases:\n%w", err)
		}

		write := !exists
		if exists && !flags[forceFlag] {
			write, err = promptYesNo("Testcases file already exists, overwrite?")
			if err != nil {
				return err
			}
		} else if exists && flags[forceFlag] {
			write = true
		}

		if write {
			printActionStart("Creating testcases")
			if err := s.WriteTestcases(p.Preview, p.ExampleTestcases); err != nil {
				return fmt.Errorf("Failed to create testcases:\n%w", err)
			}
			printActionSuccess()
		} else {
			fmt.Println("Skipping testcases.")
		}
	}

	// create code snippets, unless skipped
	if !flags[noCodeFlag] {
		langs, err := resolveLanguages(args[1:], cfg.PreferredLanguages)
		if err != nil {
			return err
		}
		// make snippets for each language, prompting if they already exist
		for _, l := range langs {
			exists, err := s.SnippetExists(p.Preview, l)
			if err != nil {
				return err
			}
			write := false
			if !exists || flags[forceFlag] {
				// if the snippet doesn't exist, or if --force is set, we write it
				write = true
			} else {
				// else we ask the user
				write, err = promptYesNo(fmt.Sprintf("%s snippet already exists, overwrite?", l.Name))
				if err != nil {
					return err
				}
			}

			if write {
				printActionStart(fmt.Sprintf("Creating %s snippet", l.Name))
				if err := s.WriteSnippet(p, l); err != nil {
					return fmt.Errorf("Failed to create snippet for %s:\n%w", l.Name, err)
				}
				printActionSuccess()
			} else {
				fmt.Printf("Skipping %s snippet.\n", l.Name)
			}
		}
	}

	fmt.Printf("Scaffolded problem %d (%s)\n", p.Number, p.Title)

	// open in editor if requested
	if flags[openFlag] {
		return openInEditor(cfg.Editor, s.GetProblemDir(p.Preview))
	}
	return nil
}

// resolveLanguages determines which languages to scaffold based on flags and config.
func resolveLanguages(specified []string, preferred []string) ([]language.Language, error) {
	switch {
	case len(specified) > 0:
		return parseLanguages(specified)
	case len(preferred) == 0:
		fmt.Println("Warning: No languages specified and no defaults configured, " +
			"no code snippets will be generated. Specify languages as part of the command " +
			"(e.g., 'leet load 123 python3 go'), or configure defaults with 'leet config edit'")
		return []language.Language{}, nil
	default:
		fmt.Printf("No languages specified, using defaults: %s\n", preferred)
		return parseLanguages(preferred)
	}
}

// parseLanguages resolves a list of language slugs/names into Language structs, deduplicating
// and validating each one.
func parseLanguages(ids []string) ([]language.Language, error) {
	seen := make(map[string]struct{})
	langs := make([]language.Language, 0, len(ids))
	for _, id := range ids {
		l, err := parseLanguage(id)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[l.Slug]; dup {
			continue
		}
		seen[l.Slug] = struct{}{}
		langs = append(langs, l)
	}
	return langs, nil
}

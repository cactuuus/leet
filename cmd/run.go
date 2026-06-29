package cmd

import (
	"fmt"
	"strings"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/spf13/cobra"
)

func NewRunCmd(ctx AppContext) *cobra.Command {
    runCmd := &cobra.Command{
        Use:            "run <number|daily> <lang>",
        Short:          "Run your solution against the example test cases.",
        Long:           "Run your solution against the example test cases.\nThese testcases are stored in 'testcases-<number>.txt', inside of the problem folder, where each argument is separated by a newline, and each testcase is separated by a divider (---). You can manually edit this file, adding or removing testcases.",
        SilenceUsage:   true,
        Args:           cobra.ExactArgs(2),
        RunE:           func(cmd *cobra.Command, args []string) error {
            return runProblem(cmd, args, ctx)
        },
    }

    runCmd.Flags().BoolP("verbose", "v", false, "Show stdout from your code, and the full compile/runtime error messages if your code fails to run.")
    runCmd.Flags().BoolP("show-all", "a", false, "Print all test cases, including passing ones.")

    return runCmd
}

func runProblem(cmd *cobra.Command, args []string, ctx AppContext) error {
    // parse flags
    verbose, err := cmd.Flags().GetBool("verbose")
    if err != nil {
        return fmt.Errorf("failed to parse --verbose flag: %w", err)
    }
    showAll, err := cmd.Flags().GetBool("show-all")
    if err != nil {
        return fmt.Errorf("failed to parse --show-all flag: %w", err)
    }

    // load core early, so that if they fail we exit before doing any work
    c, s, cfg := ctx.Client(), ctx.Scaffolder(), ctx.Config()

    // check if the user is authenticated
    if !cfg.Credentials.IsSet() {
        return fmt.Errorf("this command requires authentication. Add your LeetCode credentials in the config file using 'leet config edit'")
    }

    // get and validate language
    lang, ok := language.Get(args[1])
    if !ok {
        return fmt.Errorf("unknown language: %q — run 'leet languages' to see supported languages", args[1])
    }
    // get and validate problem
    p, err := fetchPreviewByIdentifier(c, args[0])
    if err != nil {
        return err
    }

    // check if the problem directory and the snippet file exist
    if exists, err := s.ProblemDirExists(p); err != nil {
        return fmt.Errorf("failed to check if problem %d is scaffolded: %w", p.Number, err)
    } else if !exists {
        return fmt.Errorf("problem %d is not scaffolded. Run 'leet load %d' first", p.Number, p.Number)
    }
    if exists, err := s.SnippetExists(p, lang); err != nil {
        return fmt.Errorf("failed to check if %s snippet exists for problem %d: %w", lang.Name, p.Number, err)
    } else if !exists {
        return fmt.Errorf("no %s file found for problem %d — run 'leet load %d --langs %s' first",
            lang.Name, p.Number, p.Number, lang.Slug)
    }

    // read the code snippet
    code, err := s.ReadSnippet(p, lang)
    if err != nil {
        return fmt.Errorf("failed to read solution file: %w", err)
    }

    // read the testcases
    tests, err := s.ReadTestcases(p)
    if err != nil {
        return fmt.Errorf("failed to read testcases: %w — use --input to provide test input manually", err)
    }

    // run the code
    fmt.Printf("Running problem %d (%s) in %s...", p.Number, p.Title, lang.Name)
    result, err := c.RunCode(p.Slug, p.InternalID, lang.Slug, code, tests)
    if err != nil {
        return fmt.Errorf("failed to run solution: %w", err)
    }
    fmt.Print("✓\n\n")

    printRunResult(result, tests, verbose, showAll)
    return nil
}

// printRunResult displays the result of a run in a human-readable format.
func printRunResult(r leetcode.CheckResult, inputs []string, verbose bool, showAll bool) {
    // Check if the code crashed or failed to compile
    if r.StatusCode != leetcode.ResultAccepted {
        fmt.Printf("RESULT: %s\n", r.StatusMsg)
        if r.CompileError != "" {
            fmt.Println(r.CompileError)
        }
        if r.RuntimeError != "" {
            fmt.Println(r.RuntimeError)
        }
        return
    }

    // Print the results of each testcase
    // inputs, r.ExpectedAnswer, r.CodeAnswer, and r.StdOutputList should all have the same length
    passedCount, total := 0, len(inputs)
    for i := range len(inputs) {
        in, exp, out := inputs[i], r.ExpectedAnswer[i], r.CodeAnswer[i]
        passed := exp == out
        var status string
        if passed {
            passedCount++
            if !showAll {
                continue // skip printing this testcase if it passed and showAll is false
            }
            status = "✓ PASS"
        } else {
            status = "❌FAIL" // no space cause this cross is already 2 chars wide
        }
        fmt.Printf("%s - Testcase %d\n", status, i+1)
        fmt.Printf("  Input....: (%s)\n", strings.ReplaceAll(in, "\n", ", "))
        fmt.Printf("  Expected.: %s\n", exp)
        fmt.Printf("  Output...: %s\n", out)
        if verbose {
            if r.StdOutputList[i] == "" {
                fmt.Print("  Stdout...: <no output>\n")
            } else {
                fmt.Print("  Stdout...:\n")
                lines := strings.SplitSeq(strings.Trim(r.StdOutputList[i], "\n"), "\n")
                for line := range lines {
                    fmt.Printf("\t>  %s\n", line)
                }
            }
        }
        fmt.Println()
    }
    fmt.Print("RESULT\n")
    fmt.Printf("  Testcases Passed.: %d/%d\n", passedCount, total)
    fmt.Printf("  Runtime..........: %s\n", r.StatusRuntime)
    fmt.Printf("  Memory...........: %s\n", r.StatusMemory)
}

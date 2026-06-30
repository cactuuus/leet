package cmd

import (
	"fmt"
	"strings"

	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/spf13/cobra"
)

type testCmdBoolFlag int
const (
    verboseFlag testCmdBoolFlag = iota
    showAllFlag
)

var testCmdBoolFlags = map[testCmdBoolFlag]cmdBoolFlags {
    verboseFlag: {Long: "verbose", Short: "v", Value: false, Desc: "Show stdout from your code."},
    showAllFlag: {Long: "all-tests", Short: "a", Value: false, Desc: "Print all test cases, including passing ones."},
}

func NewTestCmd(ctx AppContext) *cobra.Command {
    testCmd := &cobra.Command{
        Use:    "test <number|daily> <lang>",
        Short:  "Test your solution against the testcases defined for the problem.",
        Long:   "Test your solution against the testcases defined for the problem.\n" +
                "These testcases are stored in 'testcases-<number>.txt' (inside the problem " +
                "folder), where each argument is separated by a newline, and each testcase is " +
                "separated by a divider ('---'). You can manually edit this file, adding or " +
                "removing testcases.\n Only the failing testcases will be printed by default, " +
                "but you can use the --all-tests flag to print passing ones too.",
        Args:   cobra.ExactArgs(2),
        RunE:   func(cmd *cobra.Command, args []string) error {
            return testProblem(cmd, args, ctx)
        },
    }

    // define flags
    for _, flag := range testCmdBoolFlags {
        testCmd.Flags().BoolP(flag.Long, flag.Short, flag.Value, flag.Desc)
    }
    return testCmd
}

func testProblem(cmd *cobra.Command, args []string, ctx AppContext) error {
    // parse flags
    flags := make(map[testCmdBoolFlag]bool, len(testCmdBoolFlags))
    for id, flag := range testCmdBoolFlags {
        value, err := cmd.Flags().GetBool(flag.Long)
        if err != nil {
            return fmt.Errorf("Failed to parse --%s (-%s) flag:\n%w", flag.Long, flag.Short, err)
        }
        flags[id] = value
    }

    // load core early, so that if they fail we exit before doing any work
    cfg := ctx.Config()
    c, err := ctx.Client()
    if err != nil {
        return err
    }
    s, err := ctx.Scaffolder()
    if err != nil {
        return err
    }

    // check if the user is authenticated
	if !cfg.Credentials.IsSet() {
		return fmt.Errorf(
			"This command requires authentication. Add your LeetCode credentials in the " +
			"config file using 'leet config edit'")
	}

    // get and validate language
	l, err := parseLanguage(args[1])
	if err != nil {
		return err
	}
    // get and validate problem
    p, err := fetchPreviewByIdentifier(c, args[0])
    if err != nil {
        return err
    }

    // check if the problem directory and the snippet file exist
    if exists, err := s.ProblemDirExists(p); err != nil {
		return fmt.Errorf("Failed to check if problem %d is scaffolded:\n%w", p.Number, err)
	} else if !exists {
		return fmt.Errorf("Problem %d is not scaffolded. Run 'leet load %d' first.", p.Number, p.Number)
	}
	if exists, err := s.SnippetExists(p, l); err != nil {
		return fmt.Errorf("Failed to check if %s snippet exists for problem %d:\n%w",
			l.Name, p.Number, err)
	} else if !exists {
		return fmt.Errorf("No %s file found for problem %d. Run 'leet load %d %s' to create one.",
			l.Name, p.Number, p.Number, l.Slug)
	}

	// Read the solution file
	code, err := s.ReadSnippet(p, l)
	if err != nil {
		return fmt.Errorf("Failed to read solution file:\n%w", err)
	}

    // read the testcases
    tests, err := s.ReadTestcases(p)
    if err != nil {
        return fmt.Errorf("Failed to read testcases:\n%w", err)
    }

    // Test the code
    testMsg := fmt.Sprintf("Testing solution for %d (%s) in %s...", p.Number, p.Title, l.Name)
    printActionStart(testMsg)
    result, err := c.RunCode(p, l, code, tests)
    if err != nil {
        return fmt.Errorf("Failed to test solution:\n%w", err)
    }
    printActionSuccess()

    printTestResult(result, tests, flags[verboseFlag], flags[showAllFlag])
    return nil
}

// printTestResult displays the result of a test in a human-readable format.
func printTestResult(r leetcode.RunCheckResult, inputs []string, verbose bool, showAll bool) {
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

    if r.CorrectAnswer {
        fmt.Print("RESULT: ✓ All testcases passed\n\n")
    } else {
        fmt.Print("RESULT: ❌Some testcases failed\n\n")
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

    fmt.Printf("Testcases Passed.: %d/%d\n", passedCount, total)
    fmt.Printf("Runtime..........: %s\n", r.StatusRuntime)
    fmt.Printf("Memory...........: %s\n", r.StatusMemory)
}

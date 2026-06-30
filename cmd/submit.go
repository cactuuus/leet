package cmd

import (
	"fmt"
	"strings"

	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/spf13/cobra"
)

func NewSubmitCmd(ctx AppContext) *cobra.Command {
	submitCmd := &cobra.Command{
		Use:    "submit <number|daily> <lang>",
		Short:  "Submit your solution to LeetCode.",
		Long:	"Submit your solution to LeetCode.\nIf your solution fails, and the test case " +
				"is not already in your local testcases file, you will have the option to add it.",
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return submitProblem(cmd, args, ctx)
		},
	}

	return submitCmd
}

func submitProblem(_ *cobra.Command, args []string, ctx AppContext) error {
	cfg := ctx.Config()
	c, err := ctx.Client()
	if err != nil {
		return err
	}
	s, err := ctx.Scaffolder()
	if err != nil {
		return err
	}

	// Check authentication
	if !cfg.Credentials.IsSet() {
		return fmt.Errorf(
			"This command requires authentication. Add your LeetCode credentials in the " +
			"config file using 'leet config edit'")
	}

	// Validate language
	l, err := parseLanguage(args[1])
	if err != nil {
		return err
	}

	// Validate problem
	p, err := fetchPreviewByIdentifier(c, args[0])
	if err != nil {
		return err
	}

	// Check if the directory and file exist
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

	// Submit to LeetCode
	submitMsg := fmt.Sprintf("Submitting solution for %d (%s) in %s...", p.Number, p.Title, l.Name)
	printActionStart(submitMsg)
	result, err := c.SubmitSolution(p, l, code)
	if err != nil {
		return fmt.Errorf("Failed to submit solution:\n%w", err)
	}
	printActionSuccess()

	printSubmitResult(result)

	// If the submission failed due to a failing testcase, offer to add it to the local testcases file
	if result.StatusCode != leetcode.ResultAccepted && result.LastTestcase != "" {
		// populate existing testcases, if any
		var existingTestcases []string
		offerAdd := true
		if exists, err := s.TestcasesExists(p); err != nil {
			return fmt.Errorf("Failed to check if local testcases file exists:\n%w", err)
		} else if exists {
			existingTestcases, err = s.ReadTestcases(p)
			if err != nil {
				return fmt.Errorf("Failed to read local testcases file:\n%w", err)
			}
			for _, existing := range existingTestcases {
				if strings.TrimSpace(existing) == strings.TrimSpace(result.LastTestcase) {
					fmt.Println("\nThis testcase already exists in your local testcases file.")
					offerAdd = false
					break
				}
			}
		}
		if offerAdd {
			// prompt the user to add the failing testcase to their local testcases file
			confirm, err := promptYesNo("\nAdd this testcase to your local ones? (y/n): ")
			if err != nil {
				return fmt.Errorf("Error prompting user: %v\n", err)
			}
			if confirm {
				s.WriteTestcases(p, append(existingTestcases, result.LastTestcase))
				fmt.Println("Testcase added to your local testcases file.")
			} else {
				fmt.Println("Testcase not added.")
			}
		}
	}
	return nil
}

// printSubmitResult displays the macro outcome of an active evaluation suite.
func printSubmitResult(r leetcode.SubmitCheckResult) {
	totalTests, correctTests := "-", "-"
	if r.TotalTestcases != nil {
		totalTests = fmt.Sprintf("%d", *r.TotalTestcases)
	}
	if r.TotalCorrect != nil {
		correctTests = fmt.Sprintf("%d", *r.TotalCorrect)
	}

	if r.StatusCode == leetcode.ResultAccepted {
		fmt.Print("RESULT: ✓ Accepted\n\n")
		fmt.Printf("Testcases Passed.: %s/%s\n", correctTests, totalTests)
		runtimePercentile, memoryPercentile := "-", "-"
		if r.RuntimePercentile != nil {
			runtimePercentile = fmt.Sprintf("%.2f%%", *r.RuntimePercentile)
		}
		if r.MemoryPercentile != nil {
			memoryPercentile = fmt.Sprintf("%.2f%%", *r.MemoryPercentile)
		}
		fmt.Printf("Runtime..........: %s (beats %s)\n", r.StatusRuntime, runtimePercentile)
		fmt.Printf("Memory...........: %s (beats %s)\n", r.StatusMemory, memoryPercentile)
		return
	}

	// For Wrong Answer, TLE, MLE, Compile Error, etc.
	fmt.Printf("RESULT: ❌%s\n\n", r.StatusMsg)

	// Display testcase progress metric if available
	if r.TotalCorrect != nil && r.TotalTestcases != nil {
		fmt.Printf("Testcases Passed.: %s/%s\n", correctTests, totalTests)
	}

	// Print compilation context if it broke early
	if r.CompileError != "" {
		fmt.Printf("Compile Error:\n%s\n", r.CompileError)
		return
	}
	// Print runtime/panic details if it threw an active exception
	if r.RuntimeError != "" {
		fmt.Printf("Runtime Error:\n%s\n", r.RuntimeError)
		return
	}

	// If it failed on a specific testcase (like Wrong Answer or Time Limit Exceeded)
	if r.LastTestcase != "" {
		fmt.Println("Failing Testcase Details:")
		// Clean up line breaks for single-line display consistency
		fmt.Printf("  Input....: (%s)\n", strings.ReplaceAll(r.LastTestcase, "\n", ", "))
		if r.ExpectedOutput != "" {
			fmt.Printf("  Expected.: %s\n", r.ExpectedOutput)
		}
		if r.CodeOutput != "" {
			fmt.Printf("  Output...: %s\n", r.CodeOutput)
		}
	}
}

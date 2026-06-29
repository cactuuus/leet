package scaffold

import (
	"fmt"
	"os"
	"strings"

	"github.com/cactuuus/leet/internal/problem"
)

// testcasesSeparator is the string that separates individual test cases in the testcases file.
const testcasesSeparator = "\n---\n"

// GetTestcasesFilename returns the filename for the testcases file of a given problem.
//
// Example: "tests-1234.txt".
func (s *Scaffolder) GetTestcasesFilename(p problem.Preview) string {
	return fmt.Sprintf("tests-%d.txt", p.Number)
}

// TestcasesExists checks if the testcases file exists for a given problem.
func (s *Scaffolder) TestcasesExists(p problem.Preview) (bool, error) {
	return s.FileOrDirExists(s.GetFullFilepath(p, s.GetTestcasesFilename(p)))
}

// WriteTestcases writes the example test cases to a file, one per block, intended to be human
// readable and editable.
func (s *Scaffolder) WriteTestcases(p problem.Preview, testcases []string) error {
	// ensure the problem directory exists
	if _, err := s.CreateProblemDir(p); err != nil {
		return err
	}
	file, err := os.Create(s.GetFullFilepath(p, s.GetTestcasesFilename(p)))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(testcases, testcasesSeparator))
	return err
}

// ReadTestcases reads the testcases file and returns its contents as a slice of strings.
// This matches the format expected by the LeetCode API.
func (s *Scaffolder) ReadTestcases(p problem.Preview) ([]string, error) {
	content, err := os.ReadFile(s.GetFullFilepath(p, s.GetTestcasesFilename(p)))
	if err != nil {
		return nil, fmt.Errorf("failed to read testcases file: %w", err)
	}
	blocks := strings.Split(string(content), testcasesSeparator)
	cases := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if trimmed := strings.TrimSpace(b); trimmed != "" {
			cases = append(cases, trimmed)
		}
	}
	return cases, nil
}

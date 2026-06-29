package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cactuuus/leet/internal/problem"
)

// testcasesSep is the string that separates individual test cases in the testcases file.
const testcasesSep = "---"

// Scaffolder manages problem folders and files on disk.
type Scaffolder struct {
	problemsDir string
}

// NewScaffolder creates and returns a new Scaffolder instance.
// If the problems directory does not yet exist, it is created in the process, else an error is returned.
func NewScaffolder(problemsDir string) (*Scaffolder, error) {
	if err := os.MkdirAll(problemsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create problems directory: %w", err)
	}
	return &Scaffolder{problemsDir: problemsDir}, nil
}

// getProblemDir returns the full path to the directory for a given problem.
//
// Example: "/path/to/problems/1234.two-sum".
func (s *Scaffolder) GetProblemDir(p problem.Preview) string {
    return filepath.Join(s.problemsDir, fmt.Sprintf("%d.%s", p.Number, p.Slug))
}

// GetFullFilepath returns the full path to a file within a problem's directory.
//
// Example: "/path/to/problems/1234.two-sum/filename".
func (s *Scaffolder) GetFullFilepath(p problem.Preview, filename string) string {
	return filepath.Join(s.GetProblemDir(p), filename)
}

// ProblemDirExists checks if the directory for a given problem exists.
func (s *Scaffolder) ProblemDirExists(p problem.Preview) (bool, error) {
	return s.FileOrDirExists(s.GetProblemDir(p))
}

// SnippetExists checks if a code snippet file exists for a given problem and language.
func (s *Scaffolder) CreateProblemDir(p problem.Preview) (bool, error) {
	if err := os.MkdirAll(s.GetProblemDir(p), 0755); err != nil {
		return false, fmt.Errorf("failed to create problem directory: %w", err)
	}
	return true, nil
}

// FileOrDirExists checks if a file or directory exists at the given path.
func (s *Scaffolder) FileOrDirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

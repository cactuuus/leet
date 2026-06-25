package scaffold

import (
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
)

// Scaffolder manages problem folders and files on disk.
type Scaffolder struct {
	problemsDir string
}

func NewScaffolder(problemsDir string) (*Scaffolder, error) {
	if err := os.MkdirAll(problemsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create problems directory: %w", err)
	}
	return &Scaffolder{
		problemsDir: problemsDir,
	}, nil
}

// getProblemDir returns the full path to the directory for a given problem.
//
// Example: "/path/to/problems/1234.two-sum".
func (s *Scaffolder) GetProblemDir(p leetcode.Problem) string {
    return filepath.Join(s.problemsDir, fmt.Sprintf("%d.%s", p.Number, p.Slug))
}

// getFilename returns the filename for a given problem and language.
//
// Example: "1234.py".
func (s *Scaffolder) GetSnippetFilename(p leetcode.Problem, l language.Language) string {
	return fmt.Sprintf("%d%s", p.Number, l.Extension)
}

// getFilepath returns the full path to the file for a given problem and language.
//
// Example: "/path/to/problems/1234.two-sum/1234.py".
func (s *Scaffolder) GetSnippetFilepath(p leetcode.Problem, l language.Language) string {
	return filepath.Join(s.GetProblemDir(p), s.GetSnippetFilename(p, l))
}

// getDescFilename returns the filename for the HTML description of a given problem.
//
// Example: "desc-1234.html".
func (s *Scaffolder) GetDescFilename(p leetcode.Problem) string {
	return fmt.Sprintf("desc-%d.html", p.Number)
}

// getDescFilepath returns the full path to the HTML description file for a given problem.
//
// Example: "/path/to/problems/1234.two-sum/desc-1234.html".
func (s *Scaffolder) GetDescFilepath(p leetcode.Problem) string {
	return filepath.Join(s.GetProblemDir(p), s.GetDescFilename(p))
}

// GetProblemDirByNumber searches the problems directory for a folder belonging to the given problem
// number, without needing to fetch the problem from the API.
// Returns an error if no matching folder exists.
func (s *Scaffolder) GetProblemDirByNumber(number int) (string, error) {
	entries, err := os.ReadDir(s.problemsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read problems directory: %w", err)
	}

	prefix := fmt.Sprintf("%d.", number)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(s.problemsDir, entry.Name()), nil
		}
	}
	return "", fmt.Errorf("problem %d hasn't been loaded yet", number)
}

// SnippetExists checks if a code snippet file exists for the given problem and language.
func (s *Scaffolder) SnippetExists(p leetcode.Problem, l language.Language) (bool, error) {
	_, err := os.Stat(s.GetSnippetFilepath(p, l))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// CreateSnippet creates (or overwrites) the code file for the given problem and language.
func (s *Scaffolder) CreateSnippet(p leetcode.Problem, l language.Language) error {
	// get the code snippet
	snippet, ok := p.Snippets[l.Slug]
	if !ok {
		return fmt.Errorf("no snippet found for language %s", l.Name)
	}
	// ensures the problem directory exists
	if err := os.MkdirAll(s.GetProblemDir(p), 0755); err != nil {
		return fmt.Errorf("failed to create problem directory: %w", err)
	}
	// create the snippet file and write to it
	file, err := os.Create(s.GetSnippetFilepath(p, l))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(snippet)
	return err
}

// CreateDescription creates a new file named "problem.html" in the problem directory, containing the HTML description of the problem.
func (s *Scaffolder) CreateDescription(p leetcode.Problem) error {
	// ensures the problem directory exists
	if err := os.MkdirAll(s.GetProblemDir(p), 0755); err != nil {
		return fmt.Errorf("failed to create problem directory: %w", err)
	}
	// create the description file and write to it
	file, err := os.Create(s.GetDescFilepath(p))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(buildDescriptionHTML(p))
	return err
}

// buildDescriptionHTML is a helper function that assembles a summary of the problem, including its
// name, difficulty, link, and the content of the problem itself, into an HTML string.
func buildDescriptionHTML(p leetcode.Problem) string {
	// TODO: use proper HTML templating instead of string concatenation
	return fmt.Sprintf(
		"<h1>\n" +
		"\t<a href=\"%s\" target=\"_blank\" rel=\"noopener noreferrer\">%d. %s</a>\n" +
		"</h1>\n" +
		"<p>Difficulty: <strong>%s</strong></p>\n" +
		"<hr>\n\n" +
		"%s\n",
		p.Link,
		p.Number,
		html.EscapeString(p.Name),
		p.Difficulty,
		p.Content,
	)
}

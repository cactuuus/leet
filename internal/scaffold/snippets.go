package scaffold

import (
	"fmt"
	"os"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/problem"
)

// getFilename returns the filename for a given problem and language.
//
// Example: "1234.py".
func (s *Scaffolder) GetSnippetFilename(p problem.Preview, l language.Language) string {
	return fmt.Sprintf("%d%s", p.Number, l.Extension)
}

// SnippetExists checks if the code snippet file exists for a given problem and language.
func (s *Scaffolder) SnippetExists(p problem.Preview, l language.Language) (bool, error) {
	return s.FileOrDirExists(s.GetFullFilepath(p, s.GetSnippetFilename(p, l)))
}

// ReadSnippet reads the code snippet file for a given problem and language.
func (s *Scaffolder) ReadSnippet(p problem.Preview, l language.Language) (string, error) {
	content, err := os.ReadFile(s.GetFullFilepath(p, s.GetSnippetFilename(p, l)))
	if err != nil {
		return "", fmt.Errorf("failed to read snippet for problem %d: %w", p.Number, err)
	}
	return string(content), nil
}

// WriteSnippet creates (or overwrites) the code file for the given problem and language.
func (s *Scaffolder) WriteSnippet(p problem.Full, l language.Language) error {
	// ensure the problem directory exists
	if _, err := s.CreateProblemDir(p.Preview); err != nil {
		return err
	}
	// get the code snippet
	snippet, ok := p.Snippets[l.Slug]
	if !ok {
		return fmt.Errorf("no snippet found for language %s", l.Name)
	}
	// create the snippet file and write to it
	file, err := os.Create(s.GetFullFilepath(p.Preview, s.GetSnippetFilename(p.Preview, l)))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(snippet)
	return err
}

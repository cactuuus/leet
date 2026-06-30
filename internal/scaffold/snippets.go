package scaffold

import (
	"fmt"
	"os"
	"strings"

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
// If markers are present in the snippet, the associated sections are stripped out.
func (s *Scaffolder) ReadSnippet(p problem.Preview, l language.Language) (string, error) {
	content, err := os.ReadFile(s.GetFullFilepath(p, s.GetSnippetFilename(p, l)))
	if err != nil {
		return "", fmt.Errorf("Failed to read snippet for problem %d:\n%w", p.Number, err)
	}
	lines := strings.Split(string(content), "\n")
	// strip out anything above the start marker, if present
	for i, line := range lines {
		if strings.Contains(line, language.TemplateStartMarker) {
			lines = lines[i+1:]
			break
		}
	}
	// strip out anything below the end marker, if present
	for i, line := range lines {
		if strings.Contains(line, language.TemplateEndMarker) {
			lines = lines[:i]
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

// WriteSnippet creates (or overwrites) the code file for the given problem and language.
// It uses the template for the language and fills it with the code snippet, replacing the
// placeholder.
func (s *Scaffolder) WriteSnippet(p problem.Full, l language.Language) error {
	// ensure the problem directory exists
	if err := s.CreateProblemDir(p.Preview); err != nil {
		return err
	}
	// parse the template
	content, err := s.parseTemplate(p, l)
	if err != nil {
		return err
	}
	// finally create the snippet file and write the content to it
	file, err := os.Create(s.GetFullFilepath(p.Preview, s.GetSnippetFilename(p.Preview, l)))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

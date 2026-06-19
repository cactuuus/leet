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

// Config is the configuration for the scaffold package.
type Config interface {
	// ProblemsPath returns the path to the problems directory.
	ProblemsPath() string
}

var config Config

// Init initializes the scaffold package with the provided configuration.
func Init(cfg Config) error {
	config = cfg
	if err := os.MkdirAll(config.ProblemsPath(), 0755); err != nil {
		return fmt.Errorf("failed to create problems directory: %w", err)
	}
	return nil
}

// CheckConflicts returns languages that already have files in the problem directory.
func CheckConflicts(problem leetcode.Problem, langs []language.Language) ([]language.Language, error) {
	// if no languages are provided, there are no conflicts.
	if len(langs) == 0 {
		return []language.Language{}, nil
	}
	// if the problem directory doesn't exists, there are no conflicts in the first place.
	_, err := os.Stat(GetProblemDir(problem))
	if errors.Is(err, os.ErrNotExist) {
		return []language.Language{}, nil
	}
	if err != nil {
		return []language.Language{}, err
	}
	// check on a per-language basis.
	var conflicts []language.Language
	for _, l := range langs {
		_, err := os.Stat(GetFilepath(problem, l))
		if err == nil {
			conflicts = append(conflicts, l)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return []language.Language{}, err
		}
	}
	return conflicts, nil
}

// ScaffoldProblem creates a new problem directory (if not existing) with the provided problem and languages.
// If a file for a language already exists, it will be overwritten.
func ScaffoldProblem(problem leetcode.Problem, langs []language.Language) error {
	// create problem folder
	if err := os.MkdirAll(GetProblemDir(problem), 0755); err != nil {
		return fmt.Errorf("failed to create problem directory: %w", err)
	}
	// create description file
	if err := CreateDescription(problem); err != nil {
		return err
	}
	// create a file for each language
	for _, l := range langs {
		if err := createSnippet(problem, l); err != nil {
			return err
		}
	}
	return nil
}

// getProblemDir returns the full path to the directory for a given problem.
//
// Example: "/path/to/problems/1234.Two Sum".
func GetProblemDir(problem leetcode.Problem) string {
    return filepath.Join(config.ProblemsPath(), fmt.Sprintf("%d.%s", problem.Number, problem.Slug))
}

// getFilename returns the filename for a given problem and language.
//
// Example: "1234.py".
func GetFilename(problem leetcode.Problem, lang language.Language) string {
	return fmt.Sprintf("%d%s", problem.Number, lang.Extension)
}

// getFilepath returns the full path to the file for a given problem and language.
//
// Example: "/path/to/problems/1234.Two Sum/1234.py".
func GetFilepath(problem leetcode.Problem, lang language.Language) string {
	return filepath.Join(GetProblemDir(problem), GetFilename(problem, lang))
}

// GetProblemDirByNumber searches the problems directory for a folder belonging
// to the given problem number, without needing to fetch the problem from the API.
// Returns an error if no matching folder exists.
func GetProblemDirByNumber(number int) (string, error) {
	entries, err := os.ReadDir(config.ProblemsPath())
	if err != nil {
		return "", fmt.Errorf("failed to read problems directory: %w", err)
	}

	prefix := fmt.Sprintf("%d.", number)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(config.ProblemsPath(), entry.Name()), nil
		}
	}
	return "", fmt.Errorf("problem %d hasn't been loaded yet", number)
}

// createSnippet creates a new file for the given problem and language with the corresponding code snippet.
func createSnippet(problem leetcode.Problem, lang language.Language) error {
	file, err := os.Create(GetFilepath(problem, lang))
	if err != nil {
		return err
	}
	defer file.Close()
	// write the code snippet for the given language to the file
	snippet, ok := problem.Snippets[lang.Slug]
	if !ok {
		return fmt.Errorf("no snippet found for language %s", lang.Name)
	}
	_, err = file.WriteString(snippet)
	return err
}

// BuildDescriptionHTML assembles a summary of the problem, including its name, difficulty, link, and the content of the problem itself, into an HTML string.
func BuildDescriptionHTML(p leetcode.Problem) string {
	return fmt.Sprintf(
		`<h1><a href="%s" target="_blank" rel="noopener noreferrer">%d. %s</a></h1>
<p>Difficulty: <strong>%s</strong></p>
<hr>

%s
`,
		p.Link(),
		p.Number,
		html.EscapeString(p.Name),
		p.Difficulty,
		p.Content,
	)
}

// CreateDescription creates a new file named "problem.html" in the problem directory, containing the HTML description of the problem.
func CreateDescription(p leetcode.Problem) error {
	file, err := os.Create(filepath.Join(GetProblemDir(p), "problem.html"))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(BuildDescriptionHTML(p))
	return err
}

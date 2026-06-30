package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/problem"
)

// GetTemplateFilename returns the filename of the template file for a given language.
//
// Example: "python3.template".
func (s *Scaffolder) GetTemplateFilename(l language.Language) string {
	return fmt.Sprintf("%s.template", l.Slug)
}

// GetTemplatePath returns the full path to the template file for a given language.
//
// Example: "/path/to/templates/python3.template".
func (s *Scaffolder) GetTemplatePath(l language.Language) string {
	return filepath.Join(s.templatesDir, s.GetTemplateFilename(l))
}

// TemplateExists checks if a template file exists for a given language.
func (s *Scaffolder) TemplateExists(l language.Language) (bool, error) {
	return s.FileOrDirExists(s.GetTemplatePath(l))
}

// GetTemplate returns the template for a given language, either from the templates directory or the
//  default.
func (s *Scaffolder) GetTemplate(l language.Language) (string, error) {
	// return custom template if it exists
	if tmpl, err := os.ReadFile(s.GetTemplatePath(l)); err == nil {
		return string(tmpl), nil
	}
	// return default template otherwise
	return s.getDefaultTemplateOrBlank(l)
}

// WriteCustomTemplate makes a copy of the default template for a given language, in the custom
// templates directory.
func (s *Scaffolder) WriteCustomTemplate(l language.Language) error {
	content, err := s.getDefaultTemplateOrBlank(l)
	if err != nil {
		return err
	}
	// create the new custom template
	file, err := os.Create(s.GetTemplatePath(l))
	if err != nil {
		return fmt.Errorf("Failed to create custom template for %s:\n%w", l.Name, err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

// getDefaultTemplateOrBlank returns the default template for a given language, or an empty template
// if none is found.
func (s *Scaffolder) getDefaultTemplateOrBlank(l language.Language) (string, error) {
	// return default template otherwise
	if tmpl, err := language.GetDefaultTemplate(l); err == nil {
		return tmpl, nil
	}
	// last resort: return an empty template (containing the required placeholder(s)) and print a warning
	fmt.Printf("Warning: No default template found for language %s. Using empty template.\n", l.Name)
	return "{{.CodeSnippet}}", nil
}

// parseTemplate parses the template for a given language and fills it with the code snippet
// for a given problem, replacing the placeholder.
func (s *Scaffolder) parseTemplate(p problem.Full, l language.Language) (string, error) {
	// get the template for the language
	tmplContent, err := s.GetTemplate(l)
	if err != nil {
		return "", err
	}
	// get the code snippet
	snippet, ok := p.Snippets[l.Slug]
	if !ok {
		return "", fmt.Errorf("No snippet found for language %s", l.Name)
	}
	// parse the template and execute it with the code snippet
	tmpl, err := template.New("snippet").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("Failed to parse template for language %s:\n%w", l.Name, err)
	}
	var buf bytes.Buffer
	data := struct { CodeSnippet string }{ CodeSnippet: snippet }
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("Failed to execute template for language %s:\n%w", l.Name, err)
	}
	return buf.String(), nil
}

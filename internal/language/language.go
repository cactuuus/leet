package language

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed templates/*.template
var templates embed.FS

// these markers are used to identify the boundaries for the code snippet in the template files,
// allowing to extract the relevant code section for each language.
// IMPORTANT: these must match the markers used in the default templates.
const (
	TemplateStartMarker = "@lc-start"
	TemplateEndMarker   = "@lc-end"
)

// Language represents a programming language supported by the program, including its name, slug,
// and file extension.
type Language struct {
	Name 		string
	Slug 		string
	Extension 	string
}

// Known maps language slugs to their corresponding Language struct, listing all recognised
// languages by the application,  matching the languages supported by LeetCode and the slugs/names
// used by its API.
//
// It should be treated as a constant and not be modified at runtime.
var	known = map[string]Language{
	"cpp": 			{Name: "C++", 			Slug: "cpp", 		Extension: ".cpp"},
	"java": 		{Name: "Java", 			Slug: "java", 		Extension: ".java"},
	"python3": 		{Name: "Python3", 		Slug: "python3", 	Extension: ".py"},
	"python": 		{Name: "Python", 		Slug: "python", 	Extension: ".py"},
	"javascript": 	{Name: "JavaScript",	Slug: "javascript", Extension: ".js"},
	"typescript": 	{Name: "TypeScript",	Slug: "typescript", Extension: ".ts"},
	"csharp": 		{Name: "C#",			Slug: "csharp", 	Extension: ".cs"},
	"c": 			{Name: "C",				Slug: "c", 			Extension: ".c"},
	"golang": 		{Name: "Go",			Slug: "golang", 	Extension: ".go"},
	"kotlin": 		{Name: "Kotlin",		Slug: "kotlin", 	Extension: ".kt"},
	"swift": 		{Name: "Swift",			Slug: "swift", 		Extension: ".swift"},
	"rust": 		{Name: "Rust",			Slug: "rust", 		Extension: ".rs"},
	"ruby": 		{Name: "Ruby", 			Slug: "ruby", 		Extension: ".rb"},
	"php": 			{Name: "PHP", 			Slug: "php", 		Extension: ".php"},
	"dart": 		{Name: "Dart",			Slug: "dart", 		Extension: ".dart"},
	"scala": 		{Name: "Scala",			Slug: "scala", 		Extension: ".scala"},
	"elixir": 		{Name: "Elixir",		Slug: "elixir", 	Extension: ".ex"},
	"erlang": 		{Name: "Erlang",		Slug: "erlang", 	Extension: ".erl"},
	"racket": 		{Name: "Racket",		Slug: "racket", 	Extension: ".rkt"},
}

// Get looks up a language by slug or display name (case-insensitive).
// Returns the Language and true if found, or an empty Language and false if not.
func Get(id string) (Language, bool) {
	// normalize to lowercase for case-insensitive comparison
	id = strings.ToLower(id)
	// try slug lookup first
	if l, ok := known[id]; ok {
		return l, true
	}
	// fallback to name lookup
	for _, l := range known {
		if strings.ToLower(l.Name) == id {
			return l, true
		}
	}
	return Language{}, false
}

// All returns a slice of all known languages.
func All() []Language {
	langs := make([]Language, 0, len(known))
	for _, l := range known {
		langs = append(langs, l)
	}
	return langs
}


// GetDefaultTemplate returns the default template content for a given language.
func GetDefaultTemplate(l Language) (string, error) {
	path := filepath.Join("templates", fmt.Sprintf("%s.template", l.Slug))
	content, err := templates.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("No default template for %s:\n%w", l.Slug, err)
	}
	return string(content), nil
}

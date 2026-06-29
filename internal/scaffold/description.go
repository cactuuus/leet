package scaffold

import (
	"fmt"
	"html"
	"os"

	"github.com/cactuuus/leet/internal/problem"
)

// getDescFilename returns the filename for the HTML description of a given problem.
//
// Example: "desc-1234.html".
func (s *Scaffolder) GetDescFilename(p problem.Preview) string {
	return fmt.Sprintf("desc-%d.html", p.Number)
}

// WriteDescription creates the HTML description file for a given problem.
func (s *Scaffolder) WriteDescription(p problem.Full) error {
	// ensure the problem directory exists
	if _, err := s.CreateProblemDir(p.Preview); err != nil {
		return err
	}
	// create the description file and write to it
	file, err := os.Create(s.GetFullFilepath(p.Preview, s.GetDescFilename(p.Preview)))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(buildDescriptionHTML(p))
	return err
}

// buildDescriptionHTML is a helper function that assembles a summary of the problem, including its
// name, difficulty, link, and the content of the problem itself, into an HTML string.
func buildDescriptionHTML(p problem.Full) string {
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
		html.EscapeString(p.Title),
		p.Difficulty,
		p.Content,
	)
}

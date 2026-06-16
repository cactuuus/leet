package leetcode

import (
	"fmt"
	"regexp"
	"strings"
)

// Problem represents a LeetCode problem and its per-language starter snippets of code.
type Problem struct {
	Number		int
	Slug		string				// ID, used by the leetcode API
	Name		string
	Content		string				// raw HTML content of the problem description
	Difficulty	string
	IsPaid		bool
	Snippets	map[string]string	// language-slug -> code-snippet
}

var htmlTagRegex = regexp.MustCompile("<[^>]+>")

// Summary strips content of HTML tags and returns a short summary of the problem description.
// This can be optionally truncated to a specified length, -1 for no truncation.
func (p Problem) Summary(truncate int) string {
	// strip HTML tags
	summary := htmlTagRegex.ReplaceAllString(p.Content, "")
	// collapse whitespace
	summary = strings.Join(strings.Fields(summary), " ")
	// truncate if needed
	if truncate != -1 && len(summary) > truncate {
		summary = summary[:truncate] + "..."
	}
	return summary
}

// Link returns the URL to the problem on LeetCode.
func (p Problem) Link() string {
	return fmt.Sprintf("https://leetcode.com/problems/%s", p.Slug)
}

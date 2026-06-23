package leetcode

import (
	"fmt"
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
	Link		string				// URL to the problem
}

// problemLink constructs the URL to the problem.
func problemLink(baseURL string, slug string) string {
	return fmt.Sprintf("%s/problems/%s", baseURL, slug)
}

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
}

// Link returns the URL to the problem on LeetCode.
func (p Problem) Link() string {
	return fmt.Sprintf("https://leetcode.com/problems/%s", p.Slug)
}

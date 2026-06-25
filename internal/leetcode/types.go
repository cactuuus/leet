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

// IsEqual compares two Problem instances for equality, ignoring the Link field.
func (p Problem) IsEqual(other Problem) bool {
	if p.Number != other.Number {
		return false
	}
	if p.Slug != other.Slug {
		return false
	}
	if p.Name != other.Name {
		return false
	}
	if p.Content != other.Content {
		return false
	}
	if p.Difficulty != other.Difficulty {
		return false
	}
	if p.IsPaid != other.IsPaid {
		return false
	}
	if len(p.Snippets) != len(other.Snippets) {
		return false
	}
	for k, v := range p.Snippets {
		if other.Snippets[k] != v {
			return false
		}
	}
	// Note: We don't compare the Link field because it can be derived. It is not part of the core
	// problem data, even though it should be consistent with the other fields.
	return true
}

package problem

import (
	"maps"
	"slices"
)

// Preview represents a LeetCode problem's basic information.
// It is a lightweight struct that can be used for caching essential data.
type Preview struct {
	Number     int    `json:"number"` 		// the 'public' problem number as seen on LeetCode.
	InternalID int    `json:"internal_id"` 	// a 'private' number ID, used by the leetcode API.
	Slug       string `json:"slug"` 		// ID (usually the problem name in kebab-case).
	Title      string `json:"title"` 		// the title of the problem.
	Difficulty string `json:"difficulty"` 	// the difficulty level.
	IsPaid     bool   `json:"is_paid"` 		// indicates whether the problem is behind a paywall.
	Link	   string `json:"link"` 		// the URL to the problem.
}

// Full represents a complete LeetCode problem, embedding/extending Preview with additional details.
type Full struct {
	Preview                      		// Embedded struct
	Content    		 string            	// the problem description as raw HTML.
	Snippets   		 map[string]string 	// a map of language slugs to code snippets.
	ExampleTestcases []string      		// a list of example testcases for the problem.
}

// IsEqual compares two Full problem instances for equality by comparing all their fields.
func (p Full) IsEqual(other Full) bool {
	return p.Preview == other.Preview &&
		p.Content == other.Content &&
		maps.Equal(p.Snippets, other.Snippets) &&
		slices.Equal(p.ExampleTestcases, other.ExampleTestcases)
}

package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// questionResponse mirrors the LeetCode GraphQL response for a single problem.
type questionResponse struct {
	Data struct {
		Question struct {
			QuestionID   string `json:"questionFrontendId"`
			Title        string `json:"title"`
			TitleSlug    string `json:"titleSlug"`
			Difficulty   string `json:"difficulty"`
			IsPaidOnly   bool   `json:"isPaidOnly"`
			Content      string `json:"content"`
			CodeSnippets []struct {
				LangSlug string `json:"langSlug"`
				Code     string `json:"code"`
			} `json:"codeSnippets"`
		} `json:"question"`
	} `json:"data"`
}

// fetchProblemBySlug fetches a single problem from the LeetCode GraphQL API.
func (c *Client) fetchProblemBySlug(slug string) (Problem, error) {
	query := `query($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionFrontendId
			title
			titleSlug
			difficulty
			isPaidOnly
			content
			codeSnippets {
				langSlug
				code
			}
		}
	}`

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": map[string]string{"titleSlug": slug},
	})
	if err != nil {
		return Problem{}, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.httpClient.Post(c.baseURL+"/graphql", "application/json", bytes.NewReader(body))
	if err != nil {
		return Problem{}, fmt.Errorf("failed to fetch problem: %w", err)
	}
	defer res.Body.Close()

	var data questionResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Problem{}, fmt.Errorf("failed to parse problem response: %w", err)
	}

	q := data.Data.Question
	number, err := strconv.Atoi(q.QuestionID)
	if err != nil {
		return Problem{}, fmt.Errorf("invalid question ID: %w", err)
	}

	snippets := make(map[string]string, len(q.CodeSnippets))
	for _, s := range q.CodeSnippets {
		snippets[s.LangSlug] = s.Code
	}

	return Problem{
		Number:     number,
		Slug:       q.TitleSlug,
		Name:       q.Title,
		Content:    q.Content,
		Difficulty: q.Difficulty,
		IsPaid:     q.IsPaidOnly,
		Snippets:   snippets,
		Link:       problemLink(c.baseURL, q.TitleSlug),
	}, nil
}

// FetchProblem resolves a problem number to a slug via cache, then fetches
// the full problem details from the LeetCode GraphQL API.
func (c *Client) FetchProblem(number int) (Problem, error) {
	cache, err := c.loadCache()
	if err != nil {
		return Problem{}, err
	}

	slug, err := slugFromCache(number, cache)
	if err != nil {
		cache, err = c.refreshCache()
		if err != nil {
			return Problem{}, err
		}
		slug, err = slugFromCache(number, cache)
		if err != nil {
			return Problem{}, fmt.Errorf("problem %d does not exist", number)
		}
	}

	p, err := c.fetchProblemBySlug(slug)
	if err != nil {
		return Problem{}, err
	}
	if p.IsPaid {
		return Problem{}, fmt.Errorf("problem %d requires a LeetCode premium subscription", number)
	}
	return p, nil
}

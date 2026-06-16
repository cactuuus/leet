package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Config defines the configuration needed by the leetcode package.
type Config interface {
	// CachePath returns the path to the cache directory.
	CachePath() string
}

var config Config

// Init initializes the leetcode package with the provided configuration,
// ensuring the cache directory exists.
func Init(cfg Config) error {
	config = cfg
	if err := os.MkdirAll(filepath.Dir(config.CachePath()), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	return nil
}

// questionResponse mirrors the LeetCode GraphQL response for a single problem.
type questionResponse struct {
	Data struct {
		Question struct {
			QuestionID   string `json:"questionFrontendId"`
			Title        string `json:"title"`
			TitleSlug    string `json:"titleSlug"`
			Difficulty   string `json:"difficulty"`
			IsPaidOnly   bool   `json:"isPaidOnly"`
			Content		 string `json:"content"`
			CodeSnippets []struct {
				LangSlug string `json:"langSlug"`
				Code     string `json:"code"`
			} `json:"codeSnippets"`
		} `json:"question"`
	} `json:"data"`
}

// dailyProblemResponse mirrors the LeetCode GraphQL response for the daily problem.
type dailyResponse struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Question struct {
				TitleSlug string `json:"titleSlug"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

// fetchProblemBySlug fetches a single problem from the LeetCode GraphQL API.
func fetchProblemBySlug(slug string) (Problem, error) {
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

	res, err := http.Post("https://leetcode.com/graphql", "application/json", bytes.NewReader(body))
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
	}, nil
}

// FetchProblem resolves a problem number to a slug via cache, then fetches
// the full problem details from the LeetCode GraphQL API.
func FetchProblem(number int) (Problem, error) {
	cache, err := loadCache()
	if err != nil {
		return Problem{}, err
	}

	slug, err := slugFromCache(number, cache)
	if err != nil {
		cache, err = refreshCache()
		if err != nil {
			return Problem{}, err
		}
		slug, err = slugFromCache(number, cache)
		if err != nil {
			return Problem{}, fmt.Errorf("problem %d does not exist", number)
		}
	}

	p, err := fetchProblemBySlug(slug)
	if err != nil {
		return Problem{}, err
	}
	if p.IsPaid {
		return Problem{}, fmt.Errorf("problem %d requires a LeetCode premium subscription", number)
	}
	return p, nil
}

func FetchDailyProblem() (Problem, error) {
	query := `query {
		activeDailyCodingChallengeQuestion {
			question {
				titleSlug
			}
		}
	}`

	body, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		return Problem{}, fmt.Errorf("failed to build daily problem request: %w", err)
	}

	res, err := http.Post("https://leetcode.com/graphql", "application/json", bytes.NewReader(body))
	if err != nil {
		return Problem{}, fmt.Errorf("failed to fetch daily problem: %w", err)
	}
	defer res.Body.Close()

	var data dailyResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Problem{}, fmt.Errorf("failed to parse daily problem response: %w", err)
	}

	slug := data.Data.ActiveDailyCodingChallengeQuestion.Question.TitleSlug
	if slug == "" {
		return Problem{}, fmt.Errorf("failed to retrieve daily problem slug")
	}
	return fetchProblemBySlug(slug)
}

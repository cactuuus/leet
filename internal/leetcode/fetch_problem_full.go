package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cactuuus/leet/internal/problem"
)

type questionResponse struct {
	Number           string   `json:"questionId"`
	InternalID       string   `json:"questionFrontendId"`
	Slug             string   `json:"titleSlug"`
	Title            string   `json:"title"`
	Difficulty       string   `json:"difficulty"`
	IsPaid           bool     `json:"isPaidOnly"`
	Content          string   `json:"content"`
	ExampleTestcases []string `json:"exampleTestcaseList"`
	CodeSnippets     []struct {
		LangSlug string `json:"langSlug"`
		Code     string `json:"code"`
	} `json:"codeSnippets"`
}

type dailyResponse struct {
	Data struct {
		Daily struct {
			Date		string         		`json:"date"`
			Question 	questionResponse 	`json:"question"`
		} 									`json:"activeDailyCodingChallengeQuestion"`
	} 										`json:"data"`
}

type problemResponse struct {
	Data struct {
		Question questionResponse 	`json:"question"`
	} 								`json:"data"`
}

// fetchProblem fetches a single problem from the LeetCode GraphQL API, given its slug.
func (c *Client) fetchProblem(slug string) (problem.Full, error) {
	// GraphQL query to fetch a single problem by its slug.
	query := `query($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionId
			questionFrontendId
			title
			titleSlug
			difficulty
			isPaidOnly
			content
			exampleTestcaseList
			codeSnippets {
				langSlug
				code
			}
		}
	}`

	// Prepare the request body with the query and variables.
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": map[string]string{"titleSlug": slug},
	})
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to build request: %w", err)
	}

	// Create a POST request to the LeetCode GraphQL endpoint.
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/graphql", bytes.NewReader(body))
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("%s/problems/%s/", c.baseURL, slug))
	// Execute the request.
	res, err := c.do(req)
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to fetch problem: %w", err)
	}
	defer res.Body.Close()

	// Decode the JSON response into the questionResponse struct.
	var data problemResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return problem.Full{}, fmt.Errorf("failed to parse problem response: %w", err)
	}

    q := data.Data.Question
    internalID, err := strconv.Atoi(q.Number)
    if err != nil {
        return problem.Full{}, fmt.Errorf("invalid internal question ID: %w", err)
    }
    number, err := strconv.Atoi(q.InternalID)
    if err != nil {
        return problem.Full{}, fmt.Errorf("invalid frontend question ID: %w", err)
    }

	snippets := make(map[string]string, len(q.CodeSnippets))
	for _, s := range q.CodeSnippets {
		snippets[s.LangSlug] = s.Code
	}

	return problem.Full{
		Preview: problem.Preview{
			InternalID: internalID,
			Number:     number,
			Slug:       q.Slug,
			Title:      q.Title,
			Difficulty: q.Difficulty,
			IsPaid:     q.IsPaid,
			Link:       c.makeProblemLink(q.Slug),
		},
		Content:    		q.Content,
		Snippets:   		snippets,
		ExampleTestcases: 	q.ExampleTestcases,
	}, nil
}

// fetchDailyProblem fetches today's daily challenge from the LeetCode GraphQL API, along with its
// expiration timestamp.
func (c *Client) fetchDailyProblem() (problem.Full, int64, error) {
	// GraphQL query fetching the complete problem dataset directly from the daily challenge node.
	query := `query {
		activeDailyCodingChallengeQuestion {
			date
			question {
				questionId
				questionFrontendId
				title
				titleSlug
				difficulty
				isPaidOnly
				content
				exampleTestcaseList
				codeSnippets {
					langSlug
					code
				}
			}
		}
	}`
	// Prepare the request body with the query.
	body, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("failed to build daily problem request: %w", err)
	}

	// Create a POST request to the LeetCode GraphQL endpoint.
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/graphql", bytes.NewReader(body))
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("failed to create daily problem request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request.
	res, err := c.do(req)
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("failed to fetch daily problem: %w", err)
	}
	defer res.Body.Close()

	// Decode the JSON response into our updated dailyResponse struct.
	var data dailyResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return problem.Full{}, 0, fmt.Errorf("failed to parse daily problem response: %w", err)
	}

	q := data.Data.Daily.Question
	// Parse IDs from strings to integers
	internalID, err := strconv.Atoi(q.Number)
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("invalid internal question ID: %w", err)
	}
	number, err := strconv.Atoi(q.InternalID)
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("invalid frontend question ID: %w", err)
	}

	// Map snippets slice to lookup map
	snippets := make(map[string]string, len(q.CodeSnippets))
	for _, s := range q.CodeSnippets {
		snippets[s.LangSlug] = s.Code
	}

	// Calculate the expiration timestamp for the daily problem.
	// We use the date provided in the response to avoid timezone issues.
	// Parse it as a UTC date
	parsedDate, err := time.ParseInLocation("2006-01-02", data.Data.Daily.Date, time.UTC)
	if err != nil {
		return problem.Full{}, 0, fmt.Errorf("failed to parse daily problem date: %w", err)
	}
	// Set expiration to the exact end of that UTC day (23:59:59)
	validUntil := parsedDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second).Unix()

	return problem.Full{
		Preview: problem.Preview{
			InternalID: internalID,
			Number:     number,
			Slug:       q.Slug,
			Title:      q.Title,
			Difficulty: q.Difficulty,
			IsPaid:     q.IsPaid,
			Link:       c.makeProblemLink(q.Slug),
		},
		Content:          q.Content,
		Snippets:         snippets,
		ExampleTestcases: q.ExampleTestcases,
	}, validUntil, nil
}

func (c *Client) GetDailyProblem() (problem.Full, error) {
	// Check if the daily problem is cached.
	daily, ok, err := c.cache.GetDaily()
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to get daily problem from cache: %w", err)
	}
	if !ok {
		// cache miss, refresh and try again
		daily, validUntil, err := c.fetchDailyProblem()
		if err != nil {
			return problem.Full{}, fmt.Errorf("failed to fetch daily problem: %w", err)
		}
		// Update the cache with the new daily problem number and its expiration time.
		if err := c.cache.UpdateDaily(daily, validUntil); err != nil {
			return problem.Full{}, fmt.Errorf("failed to update daily problem in cache: %w", err)
		}
	}
	return daily, nil
}

func (c *Client) GetProblemFull(number int) (problem.Full, error) {
	// Check if the full problem is cached.
	full, ok, err := c.cache.GetFull(number)
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to get full problem from cache: %w", err)
	}
	if !ok {
		// cache miss, fetch from API and update cache
		preview, err := c.GetProblemPreview(number)
		if err != nil {
			return problem.Full{}, fmt.Errorf("failed to get problem preview: %w", err)
		}
		// Fetch the full problem details using the slug from the preview.
		full, err = c.fetchProblem(preview.Slug)
		if err != nil {
			return problem.Full{}, fmt.Errorf("failed to fetch problem: %w", err)
		}
		if err := c.cache.UpdateFull(full); err != nil {
			return problem.Full{}, fmt.Errorf("failed to update full problem in cache: %w", err)
		}
	}
	return full, nil
}

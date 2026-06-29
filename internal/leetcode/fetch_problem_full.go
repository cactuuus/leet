package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cactuuus/leet/internal/problem"
)

// dailyResponse mirrors the LeetCode GraphQL response for the daily problem.
type dailyResponse struct {
	Data struct {
		Daily struct {
			Question struct {
				Slug string `json:"titleSlug"`
			} 				`json:"question"`
		} 					`json:"activeDailyCodingChallengeQuestion"`
	} 						`json:"data"`
}

// questionResponse mirrors the LeetCode GraphQL response for a single problem.
type questionResponse struct {
	Data struct {
		Question struct {
            Number       		string   	`json:"questionId"`
            InternalID   		string   	`json:"questionFrontendId"`
            Slug         		string   	`json:"titleSlug"`
            Title        		string   	`json:"title"`
            Difficulty   		string   	`json:"difficulty"`
            IsPaid       		bool     	`json:"isPaidOnly"`
            Content      		string   	`json:"content"`
            ExampleTestcases 	[]string 	`json:"exampleTestcaseList"`
			CodeSnippets 		[]struct {
				LangSlug 		string 		`json:"langSlug"`
				Code     		string 		`json:"code"`
			} 								`json:"codeSnippets"`
		} 									`json:"question"`
	} 										`json:"data"`
}

// fetchProblemBySlug fetches a single problem from the LeetCode GraphQL API.
func (c *Client) fetchProblemBySlug(slug string) (problem.Full, error) {
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
	var data questionResponse
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

// FetchProblem fetches the full problem details from the LeetCode GraphQL API, using the problem
// number.
// It first fetches the problem preview to get the slug (likly cached), then uses that slug to fetch
// the full problem.
func (c *Client) FetchProblem(number int) (problem.Full, error) {
	// Get preview first, which contains the slug needed to fetch the full problem.
	// This might trigger a cache refresh, meaning an additional API call to fetch all previews.
	preview, err := c.GetProblemPreview(number)
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to get problem preview: %w", err)
	}

	// Fetch the full problem details using the slug from the preview.
	full, err := c.fetchProblemBySlug(preview.Slug)
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to fetch problem: %w", err)
	}
	// TODO: a user using credentials might be able to fetch a problem that is behind a paywall. We
	// should check if the user has access to the problem and return an error if not.
	if full.IsPaid {
		return problem.Full{}, fmt.Errorf("problem %d requires a LeetCode premium subscription", number)
	}
	return full, nil
}


// FetchDailyProblem fetches today's daily challenge from the LeetCode GraphQL API.
// It first fetches the daily problem's slug, then uses that slug to fetch the full problem.
func (c *Client) FetchDailyProblem() (problem.Full, error) {
	// GraphQL query to fetch the daily problem's slug.
	query := `query {
		activeDailyCodingChallengeQuestion {
			question {
				titleSlug
			}
		}
	}`

	// Prepare the request body with the query.
	body, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to build daily problem request: %w", err)
	}

	// Create a POST request to the LeetCode GraphQL endpoint.
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/graphql", bytes.NewReader(body))
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to create daily problem request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Execute the request.
	res, err := c.do(req)
	if err != nil {
		return problem.Full{}, fmt.Errorf("failed to fetch daily problem: %w", err)
	}
	defer res.Body.Close()

	// Decode the JSON response into the dailyResponse struct.
	var data dailyResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return problem.Full{}, fmt.Errorf("failed to parse daily problem response: %w", err)
	}

	// Extract the slug from the response and fetch the full problem using that slug.
	slug := data.Data.Daily.Question.Slug
	if slug == "" {
		return problem.Full{}, fmt.Errorf("failed to retrieve daily problem slug")
	}
	return c.fetchProblemBySlug(slug)
}

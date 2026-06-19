package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// dailyResponse mirrors the LeetCode GraphQL response for the daily problem.
type dailyResponse struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Question struct {
				TitleSlug string `json:"titleSlug"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

// FetchDailyProblem fetches today's daily challenge from the LeetCode GraphQL API.
func (c *Client) FetchDailyProblem() (Problem, error) {
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

	res, err := c.httpClient.Post(c.baseURL+"/graphql", "application/json", bytes.NewReader(body))
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
	return c.fetchProblemBySlug(slug)
}

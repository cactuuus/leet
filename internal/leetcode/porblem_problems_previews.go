package leetcode

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cactuuus/leet/internal/problem"
)

// fetchPreviewsResponse represents the structure of the response from the LeetCode API when
// fetching all problems, keeping only the essential fields we need for problem previews.
type fetchPreviewsResponse struct {
	StatStatusPairs []struct {
		Stat struct {
			Number  	int    	`json:"frontend_question_id"`
			InternalID  int    	`json:"question_id"`
			Slug        string 	`json:"question__title_slug"`
			Title       string 	`json:"question__title"`
		} 						`json:"stat"`
		Difficulty struct {
			Level int 			`json:"level"`
		} 						`json:"difficulty"`
		IsPaid bool   			`json:"paid_only"`
	}							`json:"stat_status_pairs"`
}

// FetchPreviews fetches all problem previews from the LeetCode API and returns them as a slice of
// problem.Preview.
func (c *Client) FetchPreviews() ([]problem.Preview, error) {
	// Create a GET request to the LeetCode API endpoint for all problems.
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/problems/all/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for all problems: %w", err)
	}

	// Execute the request.
	res, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch problems: %w", err)
	}
	defer res.Body.Close()

	// Decode the JSON response into the fetchPreviewsResponse struct.
	var data fetchPreviewsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse problems response: %w", err)
	}

	// Convert the response data into a slice of problem.Preview.
	var previews []problem.Preview
	for _, pair := range data.StatStatusPairs {
		previews = append(previews, problem.Preview{
			Number:     pair.Stat.Number,
			InternalID: pair.Stat.InternalID,
			Slug:       pair.Stat.Slug,
			Title:      pair.Stat.Title,
			Difficulty: levelToDifficulty(pair.Difficulty.Level),
			IsPaid:     pair.IsPaid,
			Link:       c.makeProblemLink(pair.Stat.Slug),
		})
	}
	return previews, nil
}

// levelToDifficulty converts a numeric difficulty level to its string representation.
func levelToDifficulty(level int) string {
	switch level {
	case 1:
		return "Easy"
	case 2:
		return "Medium"
	case 3:
		return "Hard"
	default:
		return "Unknown"
	}
}


// GetProblemPreview returns the preview for a given problem number.
// If the preview is not in the cache, it refreshes the cache and tries once again.
func (c *Client) GetProblemPreview(number int) (problem.Preview, error) {
	preview, ok, err := c.cache.GetPreview(number)
	if err != nil {
		return problem.Preview{}, fmt.Errorf("failed to get preview from cache: %w", err)
	}
	if !ok {
		// cache miss, refresh and try again
		previews, err := c.FetchPreviews()
		if err != nil {
			return problem.Preview{}, fmt.Errorf("failed to fetch previews: %w", err)
		}
		c.cache.UpdatePreviews(previews...)
		if err := c.cache.Save(); err != nil {
			return problem.Preview{}, fmt.Errorf("failed to save cache: %w", err)
		}
		// if the slug is still not in the cache, it means the problem number is invalid
		preview, ok, err = c.cache.GetPreview(number)
		if err != nil {
			return problem.Preview{}, fmt.Errorf("failed to get preview from cache: %w", err)
		}
		if !ok {
			return problem.Preview{}, fmt.Errorf("Problem %d doesn't exist", number)
		}
	}
	return preview, nil
}

// Example of an entry in the response from the LeetCode API when fetching all problems.
//   {
//     "stat": {
//       "question_id": 4342,
//       "question__article__live": null,
//       "question__article__slug": null,
//       "question__article__has_video_solution": null,
//       "question__title": "Create Grid With Exactly One Path",
//       "question__title_slug": "create-grid-with-exactly-one-path",
//       "question__hide": false,
//       "total_acs": 37075,
//       "total_submitted": 48851,
//       "frontend_question_id": 3963,
//       "is_new_question": false
//     },
//     "status": null,
//     "difficulty": {
//       "level": 1
//     },
//     "paid_only": false,
//     "is_favor": false,
//     "frequency": 0,
//     "progress": 0
//   },

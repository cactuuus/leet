package leetcode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// problemsResponse mirrors the LeetCode REST API response for all problems.
type problemsResponse struct {
	StatStatusPairs []struct {
		Stat struct {
			QuestionID  int    `json:"frontend_question_id"`
			TitleSlug   string `json:"question__title_slug"`
		} `json:"stat"`
	} `json:"stat_status_pairs"`
}

// loadCache loads the problem slug map from disk.
// Returns an empty map if the cache file does not exist yet.
func (c *Client) loadCache() (map[int]string, error) {
	file, err := os.Open(c.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		return map[int]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var cache map[int]string
	if err := json.NewDecoder(file).Decode(&cache); err != nil {
		return nil, fmt.Errorf("failed to decode cache file: %w", err)
	}
	return cache, nil
}

// refreshCache fetches all problem slugs from the LeetCode API, persists them to disk,
// and returns the updated map.
func (c *Client) refreshCache() (map[int]string, error) {
	slugs, err := c.fetchAllSlugs()
	if err != nil {
		return nil, err
	}
	if err := c.saveCache(slugs); err != nil {
		return nil, err
	}
	return slugs, nil
}

// saveCache saves the problem slug map to disk.
func (c *Client) saveCache(cache map[int]string) error {
	file, err := os.Create(c.cachePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(cache); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

// fetchAllSlugs fetches all problem slugs from the LeetCode REST API.
func (c *Client) fetchAllSlugs() (map[int]string, error) {
	res, err := c.httpClient.Get(c.baseURL + "/api/problems/all/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch problems: %w", err)
	}
	defer res.Body.Close()

	var data problemsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse problems response: %w", err)
	}

	slugs := make(map[int]string, len(data.StatStatusPairs))
	for _, p := range data.StatStatusPairs {
		slugs[p.Stat.QuestionID] = p.Stat.TitleSlug
	}
	return slugs, nil
}

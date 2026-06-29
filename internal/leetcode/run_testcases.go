package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	pollInterval = time.Second
	pollTimeout  = 30 * time.Second
)

// runTestsRequest mirrors the body expected by LeetCode's interpret_solution endpoint.
type runTestsRequest struct {
	Lang       string `json:"lang"`
	QuestionID int    `json:"question_id"`
	TypedCode  string `json:"typed_code"`
	DataInput  string `json:"data_input"`
}

// interpretResponse mirrors the response from LeetCode's interpret_solution endpoint.
type interpretResponse struct {
	InterpretID string `json:"interpret_id"`
}

// RunCode submits code to LeetCode's interpret endpoint and polls for the result.
// dataInput is the test input to run against; an empty string uses LeetCode's first example.
func (c *Client) RunCode(slug string, internalID int, langSlug, code string, tests []string) (CheckResult, error) {
	body, err := json.Marshal(runTestsRequest{
		Lang:       langSlug,
		QuestionID: internalID,
		TypedCode:  code,
		DataInput:  strings.Join(tests, "\n"),
	})
	if err != nil {
		return CheckResult{}, fmt.Errorf("failed to build run_test request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/problems/%s/interpret_solution/", c.baseURL, slug),
		bytes.NewReader(body),
	)
	if err != nil {
		return CheckResult{}, fmt.Errorf("failed to create run-tests request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("%s/problems/%s/", c.baseURL, slug))

	res, err := c.do(req)
	if err != nil {
		return CheckResult{}, fmt.Errorf("failed to send run-tests request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return CheckResult{}, fmt.Errorf("run-tests request failed with status %s", res.Status)
	}

	var interpret interpretResponse
	if err := json.NewDecoder(res.Body).Decode(&interpret); err != nil {
		return CheckResult{}, fmt.Errorf("failed to parse run-tests response: %w", err)
	}

	return c.pollCheck(interpret.InterpretID)
}

// pollCheck polls the check endpoint until the result is ready or the timeout is reached.
func (c *Client) pollCheck(id string) (CheckResult, error) {
	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/submissions/detail/%s/check/", c.baseURL, id),
			nil,
		)
		if err != nil {
			return CheckResult{}, fmt.Errorf("failed to create poll request: %w", err)
		}

		res, err := c.do(req)
		if err != nil {
			return CheckResult{}, fmt.Errorf("failed to poll result: %w", err)
		}

		var result CheckResult
		err = json.NewDecoder(res.Body).Decode(&result)
		res.Body.Close()
		if err != nil {
			return CheckResult{}, fmt.Errorf("failed to parse poll response: %w", err)
		}

		if result.State != "STARTED" && result.State != "PENDING" {
			return result, nil
		}

		time.Sleep(pollInterval)
	}
	return CheckResult{}, fmt.Errorf("timed out waiting for result after %s", pollTimeout)
}

package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/problem"
)

// runTestsRequest mirrors the body expected by LeetCode's interpret_solution endpoint.
type runTestsRequest struct {
	Lang       string `json:"lang"`
	QuestionID int    `json:"question_id"`
	TypedCode  string `json:"typed_code"`
	DataInput  string `json:"data_input"`
}

// runTestsResponse mirrors the response from LeetCode's interpret_solution endpoint.
type runTestsResponse struct {
	InterpretID string `json:"interpret_id"`
}

// RunCode submits code to LeetCode's interpret endpoint and polls for the result.
// dataInput is the test input to run against; an empty string uses LeetCode's first example.
func (c *Client) RunCode(p problem.Preview, l language.Language, code string, tests []string) (RunCheckResult, error) {
	body, err := json.Marshal(runTestsRequest{
		Lang:       l.Slug,
		QuestionID: p.InternalID,
		TypedCode:  code,
		DataInput:  strings.Join(tests, "\n"),
	})
	if err != nil {
		return RunCheckResult{}, fmt.Errorf("failed to build run-test request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/problems/%s/interpret_solution/", c.baseURL, p.Slug),
		bytes.NewReader(body),
	)
	if err != nil {
		return RunCheckResult{}, fmt.Errorf("failed to create run-tests request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("%s/problems/%s/", c.baseURL, p.Slug))

	res, err := c.do(req)
	if err != nil {
		return RunCheckResult{}, fmt.Errorf("failed to send run-tests request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return RunCheckResult{}, fmt.Errorf("run-tests request failed with status %s", res.Status)
	}

	var runTestsRes runTestsResponse
	if err := json.NewDecoder(res.Body).Decode(&runTestsRes); err != nil {
		return RunCheckResult{}, fmt.Errorf("failed to parse run-tests response: %w", err)
	}

	var result RunCheckResult
	if err := c.pollCheck(runTestsRes.InterpretID, &result); err != nil {
		return RunCheckResult{}, fmt.Errorf("failed to poll run-tests result: %w", err)
	}

	return result, nil
}

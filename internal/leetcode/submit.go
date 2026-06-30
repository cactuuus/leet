package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/problem"
)

type submitRequest struct {
	Lang       string `json:"lang"`
	QuestionID int    `json:"question_id"`
	TypedCode  string `json:"typed_code"`
}

// submitResponse mirrors the response from LeetCode's submit endpoint.
type submitResponse struct {
	SubmissionID int `json:"submission_id"`
}

// SubmitSolution submits the solution for a given problem and language.
// It then polls for the result.
func (c *Client) SubmitSolution(p problem.Preview, l language.Language, code string) (SubmitCheckResult, error) {
	body, err := json.Marshal(submitRequest{
		Lang:       l.Slug,
		QuestionID: p.InternalID,
		TypedCode:  code,
	})
	if err != nil {
		return SubmitCheckResult{}, fmt.Errorf("Failed to build submit-solution request:\n%w", err)
	}

	endpoint, err := c.makeURL("problems", p.Slug, "submit")
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return SubmitCheckResult{}, fmt.Errorf("Failed to create submit-solution request:\n%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.do(req, p.Link)
	if err != nil {
		return SubmitCheckResult{}, fmt.Errorf("Failed to send submit-solution request:\n%w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return SubmitCheckResult{}, fmt.Errorf("Submit-solution request failed with status %s", res.Status)
	}

	var submitRes submitResponse
	if err = json.NewDecoder(res.Body).Decode(&submitRes); err != nil {
		return SubmitCheckResult{}, fmt.Errorf("Failed to parse submit-solution response:\n%w", err)
	}

	var result SubmitCheckResult
	if err := c.pollCheck(strconv.Itoa(submitRes.SubmissionID), &result); err != nil {
		return SubmitCheckResult{}, fmt.Errorf("Failed to poll submit-solution result:\n%w", err)
	}

	return result, nil
}

package leetcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// how often and how long to poll for results after a run-tests or submit operation
const (
	pollInterval = time.Second
	pollTimeout  = 30 * time.Second
)

// ResultCode represents the status of a code submission or run.
type ResultCode int
const (
	ResultAccepted      ResultCode = 10
	ResultWrongAnswer   ResultCode = 11
	ResultTimeLimit     ResultCode = 14
	ResultMemoryLimit   ResultCode = 15
	ResultCompileError  ResultCode = 20
)

// ResultState represents the state of a code submission or run.
type ResultState string
const (
	StateStarted ResultState = "STARTED"
	StatePending ResultState = "PENDING"
	StateSuccess ResultState = "SUCCESS"
)

// BaseCheckResult holds fields shared by both run and submit operations.
type BaseCheckResult struct {
	State         ResultState `json:"state"`
	RunSuccess    bool        `json:"run_success"`
	// run-test always returns "Accepted" as long as the code compiles and runs, even if the output is wrong.
	StatusMsg     string      `json:"status_msg"`
	StatusCode    ResultCode  `json:"status_code"`
	RuntimeError  string      `json:"runtime_error"`		// The runtime error message, if any
	CompileError  string      `json:"full_compile_error"`	// The full compile error message, if any
	StatusRuntime string      `json:"status_runtime"`		// e.g. "52 ms"
	StatusMemory  string      `json:"status_memory"`		// e.g. "3.5 MB"
}

// RunCheckResult holds the result of a run-tests request. It embeds BaseCheckResult.
type RunCheckResult struct {
	BaseCheckResult
	CorrectAnswer  bool     `json:"correct_answer"`			// True if CodeAnswer perfectly matches ExpectedAnswer
	CodeAnswer     []string `json:"code_answer"`			// Returned output per testcase
	ExpectedAnswer []string `json:"expected_code_answer"`	// Correct, expected output per testcase
	StdOutputList  []string `json:"std_output_list"`		// StdOut per testcase
}

// SubmitCheckResult holds the result of a submit-solution request. It embeds BaseCheckResult.
type SubmitCheckResult struct {
	BaseCheckResult
	TotalCorrect   *int   `json:"total_correct"`
	TotalTestcases *int   `json:"total_testcases"`
	LastTestcase   string `json:"last_testcase"`		// The raw input that caused the failure
	CodeOutput     string `json:"code_output"`			// What your function returned on failure
	ExpectedOutput string `json:"expected_output"`		// What it should have returned
	RuntimePercentile *float64 `json:"runtime_percentile"`
	MemoryPercentile  *float64 `json:"memory_percentile"`
}

// interface to unify RunCheckResult and SubmitCheckResult for polling
type pollableResult interface {
	GetState() ResultState
}

// GetState returns the state of the result, allowing it to be used in polling.
func (r BaseCheckResult) GetState() ResultState {
	return r.State
}

// String returns a human-readable representation of the status code.
func (s ResultCode) String() string {
	switch s {
	case ResultAccepted:
		return "Accepted"
	case ResultWrongAnswer:
		return "Wrong Answer"
	case ResultTimeLimit:
		return "Time Limit Exceeded"
	case ResultMemoryLimit:
		return "Memory Limit Exceeded"
	case ResultCompileError:
		return "Compile Error"
	default:
		return fmt.Sprintf("Unknown Status Code (%d)", int(s))
	}
}

// pollCheck polls the check endpoint until the result is ready or the timeout is reached.
func (c *Client) pollCheck(id string, target pollableResult) error {
	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		endpoint, err := c.makeURL("submissions", "detail", id, "check")
		if err != nil {
			return fmt.Errorf("Failed to build request URL for poll-check:\n%w", err)
		}
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("Failed to create poll request:\n%w", err)
		}

		res, err := c.do(req, c.baseURL)
		if err != nil {
			return fmt.Errorf("Failed to poll result:\n%w", err)
		}

		err = json.NewDecoder(res.Body).Decode(&target)
		res.Body.Close()
		if err != nil {
			return fmt.Errorf("Failed to parse poll response:\n%w", err)
		}

		if target.GetState() != StateStarted && target.GetState() != StatePending {
			return nil
		}

		time.Sleep(pollInterval)
	}
	return fmt.Errorf("Timed out waiting for result after %s seconds.", pollTimeout)
}

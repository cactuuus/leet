package leetcode

import "fmt"

// ResultCode represents the status of a code submission or run.
type ResultCode int
const (
	ResultAccepted      ResultCode = 10
	ResultWrongAnswer   ResultCode = 11
	ResultTimeLimit     ResultCode = 14
	ResultMemoryLimit   ResultCode = 15
	ResultCompileError  ResultCode = 20
)

// CheckResult holds the result of a run-tests or submit operation, as returned by the poll endpoint.
type CheckResult struct {
	State          string   `json:"state"`             // "STARTED", "PENDING", "SUCCESS"
	RunSuccess     bool     `json:"run_success"`
	StatusMsg      string   `json:"status_msg"`        // Run: "Accepted" (if no crash). Submit: "Accepted", "Wrong Answer", etc.
	StatusCode     ResultCode `json:"status_code"`     // 10=accepted, 11=wrong answer, 14=TLE, 15=MLE, 20=compile error

	// --- RUN SPECIFIC FIELDS ---
	CorrectAnswer  bool     `json:"correct_answer"`       // True if CodeAnswer perfectly matches ExpectedAnswer
	CodeAnswer     []string `json:"code_answer"`          // Your function's returned output per testcase
	ExpectedAnswer []string `json:"expected_code_answer"` // The reference correct output per testcase
	StdOutputList  []string `json:"std_output_list"`      // stdout (fmt.Println, etc) per testcase

	RuntimeError   string   `json:"runtime_error"`		  // The runtime error message, if any
	CompileError   string   `json:"full_compile_error"`	  // The full compile error message, if any
	StatusRuntime  string   `json:"status_runtime"`       // e.g. "52 ms"
	StatusMemory   string   `json:"status_memory"`        // e.g. "3.5 MB"

	// --- SUBMIT SPECIFIC FIELDS ---
	TotalCorrect   *int     `json:"total_correct"`
	TotalTestcases *int     `json:"total_testcases"`
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

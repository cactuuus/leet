package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
)

var (
	// some useful variables for testing

	inProblem = []language.Language{
		{Name: "Python3", Slug: "python3", Extension: ".py"},
		{Name: "Go", Slug: "golang", Extension: ".go"},
	}
	notInProblem = []language.Language{
		{Name: "C", Slug: "c", Extension: ".c"},
		{Name: "PHP", Slug: "php", Extension: ".php"},
	}
	testProblem = leetcode.Problem{
		Number:     1234,
		Slug:       "test-problem",
		Name:       "Test Problem",
		Content:    "<div>Test</div>",
		Difficulty: "Medium",
		IsPaid:     false,
		Snippets:   map[string]string{
			inProblem[0].Slug : "test-language-0",
			inProblem[1].Slug : "test-language-1",
	 	},
		Link:       "test-problem-link",
	}
)

// testScaffolder is a helper function that creates a new Scaffolder instance for testing, using a
// temporary directory. It should always work.
func testScaffolder(t *testing.T) *Scaffolder {
	t.Helper()
	problemsDir := filepath.Join(t.TempDir(), "problems")
	scaffolder, err := NewScaffolder(problemsDir)
	if err != nil {
		t.Fatalf("error creating new scaffolder: %v", err)
	}
	return scaffolder
}

func TestNewScaffolder_Valid(t* testing.T) {
	problemsDir := filepath.Join(t.TempDir(), "problems")
	scaffolder, err := NewScaffolder(problemsDir)
	if err != nil {
		t.Fatalf("NewScaffolder returned unexpected error: %v", err)
	}
	if scaffolder.problemsDir != problemsDir {
		t.Errorf("problemsDir = %q, want %q", scaffolder.problemsDir, problemsDir)
	}
}

func TestNewScaffolder_Invalid(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatalf("failed to create blocker file: %v", err)
	}
	_, err := NewScaffolder(blocker)
	if err == nil {
		t.Errorf("NewScaffolder(%q) expected error, got nil", blocker)
	}
}

func TestCreateSnippet_Valid(t *testing.T) {
	scaffolder := testScaffolder(t)
	for _, l := range inProblem {
		if err := scaffolder.CreateSnippet(testProblem, l); err != nil {
			t.Fatalf("CreateSnippet(%q)( unexpected error: %v", l.Slug, err)
		}
		content, err := os.ReadFile(scaffolder.GetSnippetFilepath(testProblem, l))
		if err != nil {
			t.Fatalf("Failed to read snippet file: %v", err)
		}
		wrote, expected := string(content), testProblem.Snippets[l.Slug]
		if wrote != expected {
			t.Errorf("snippet content = %q, want %q", wrote, expected)
		}
	}
}

func TestCreateSnippet_NotAvailable(t *testing.T) {
	scaffolder := testScaffolder(t)
	for _, l := range notInProblem {
		if err := scaffolder.CreateSnippet(testProblem, l); err == nil{
			t.Errorf("CreateSnippet(%q) expected error, got nil", l.Slug)
		}
	}
}

func TestCreateDescription(t *testing.T) {
	scaffolder := testScaffolder(t)
	if err := scaffolder.CreateDescription(testProblem); err != nil {
		t.Fatalf("CreateDescription() unexpected error: %v", err)
	}
	content, err := os.ReadFile(scaffolder.GetDescFilepath(testProblem))
	if err != nil {
		t.Fatalf("Failed to read description file: %v", err)
	}
	wrote, expected := string(content), buildDescriptionHTML(testProblem)
	if wrote != expected {
		t.Errorf("description content = %q, want %q", wrote, expected)
	}
}

func TestSnippetExists(t *testing.T) {
	scaffolder := testScaffolder(t)
	l := inProblem[0] // we pick a randome language that is in the problem

	// check that the snippet does not exist before creation
	exists, err := scaffolder.SnippetExists(testProblem, l)
	if err != nil {
		t.Fatalf("SnippetExists unexpected error: %v", err)
	}
	if exists {
		t.Errorf("SnippetExists() = true before creation, want false")
	}

	// check that the snippet exists after creation
	if err := scaffolder.CreateSnippet(testProblem, l); err != nil {
		t.Fatalf("CreateSnippet() unexpected error: %v", err)
	}
	exists, err = scaffolder.SnippetExists(testProblem, l)
	if err != nil {
		t.Fatalf("SnippetExists() unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("SnippetExists() = false after creation, want true")
	}
}

func TestGetProblemDirByNumber(t *testing.T) {
	scaffolder := testScaffolder(t)

	// check that the problem directory does not exist before creation
	_, err := scaffolder.GetProblemDirByNumber(testProblem.Number)
	if err == nil {
		t.Errorf("GetProblemDirByNumber() expected error before creation, got nil")
	}

	// create the problem directory by creating a description for it
	if err := scaffolder.CreateDescription(testProblem); err != nil {
		t.Fatalf("CreateDescription() unexpected error: %v", err)
	}
	dir, err := scaffolder.GetProblemDirByNumber(testProblem.Number)
	if err != nil {
		t.Fatalf("GetProblemDirByNumber() unexpected error after creation: %v", err)
	}
	expectedDir := scaffolder.GetProblemDir(testProblem)
	if dir != expectedDir {
		t.Errorf("GetProblemDirByNumber() = %q, want %q", dir, expectedDir)
	}
}

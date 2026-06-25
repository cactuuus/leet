package leetcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mock server for testing
var (
	testServer *httptest.Server
	dailyProblem = Problem{
		Number:     3,
		Slug:       "daily-problem",
		Name:       "Daily Problem",
		Content:    "<div>Test</div>",
		Difficulty: "Medium",
		IsPaid:     false,
		Snippets:   map[string]string{
			"python" : "daily-language-0",
			"golang" : "daily-language-1",
		},
	}
	paidProblem = Problem{
		Number:     4,
		Slug:       "paid-problem",
		Name:       "Paid Problem",
		Content:    "<div>Test</div>",
		Difficulty: "Hard",
		IsPaid:     true,
		Snippets:   map[string]string{
			"python" : "paid-language-0",
			"golang" : "paid-language-1",
		},
	}
	mockProblems = []Problem{
		dailyProblem,
		paidProblem,
		{
			Number:     1,
			Slug:       "two-sum",
			Name:       "Two Sum",
			Content:    "<div>Test</div>",
			Difficulty: "Easy",
			IsPaid:     false,
			Snippets:   map[string]string{
				"python" : "test-language-0",
				"golang" : "test-language-1",
			},
		},
		{
			Number:     2,
			Slug:       "add-two-numbers",
			Name:       "Add Two Numbers",
			Content:    "<div>Test</div>",
			Difficulty: "Medium",
			IsPaid:     false,
			Snippets:   map[string]string{
				"cpp" : "test-language-2",
				"c" : "test-language-3",
			},
		},
	}
)

// TestMain is the entry point for testing in this package, it sets up the mock server and runs the
// tests.
func TestMain(m *testing.M) {
	testServer = httptest.NewServer(http.HandlerFunc(handleTestRequest))
	m.Run()
	testServer.Close()
}

func mockCache() map[int]string {
	cache := make(map[int]string)
	for _, p := range mockProblems {
		cache[p.Number] = p.Slug
	}
	return cache
}

// handleTestRequest is a helper function to handle a given request, which mimics how real request
// would be handled by LeetCode.
func handleTestRequest(w http.ResponseWriter, r *http.Request) {
	switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/problems/all/":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"stat_status_pairs":[`))
			for i, p := range mockProblems {
				str := fmt.Sprintf(
					`{"stat":{"frontend_question_id":%d,"question__title_slug":"%s"}}`,
					p.Number,
					p.Slug,
				)
				w.Write([]byte(str))
				if i < len(mockProblems)-1 {
					w.Write([]byte(","))
				}
			}
			w.Write([]byte("]}"))
		case r.Method == http.MethodPost && r.URL.Path == "/graphql":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			query, _ := body["query"].(string)
			w.Header().Set("Content-Type", "application/json")

			if strings.Contains(query, "activeDailyCodingChallengeQuestion") {
				json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"activeDailyCodingChallengeQuestion": map[string]any{
							"question": map[string]any{
								"titleSlug": dailyProblem.Slug,
							},
						},
					},
				})
				return
			}

			// single problem query — extract slug from variables
			vars, _ := body["variables"].(map[string]any)
			slug, _ := vars["titleSlug"].(string)

			var found *Problem
			for i := range mockProblems {
				if mockProblems[i].Slug == slug {
					found = &mockProblems[i]
					break
				}
			}
			if found == nil {
				http.NotFound(w, r)
				return
			}

			snippets := make([]map[string]string, 0, len(found.Snippets))
			for langSlug, code := range found.Snippets {
				snippets = append(snippets, map[string]string{
					"langSlug": langSlug,
					"code":     code,
				})
			}

			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"question": map[string]any{
						"questionFrontendId": fmt.Sprintf("%d", found.Number),
						"title":             found.Name,
						"titleSlug":         found.Slug,
						"difficulty":        found.Difficulty,
						"isPaidOnly":        found.IsPaid,
						"content":           found.Content,
						"codeSnippets":      snippets,
					},
				},
			})
		default:
			http.NotFound(w, r)
	}
}

func testClient(t *testing.T) *Client {
	t.Helper()
	cachePath := t.TempDir() + "/problems.json"
	c, err := newClient(cachePath, testServer.URL, testServer.Client())
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

// TestCache tests all (implicitly or explicitly) cache-related functions.
func TestCache(t *testing.T) {
	c := testClient(t)
	// check that cache returns empty map if file doesn't exists
	cache, err := c.loadCache()
	if err != nil {
		t.Fatalf("error loading cache, before saving: %v", err)
	}
	if len(cache) != 0 {
		t.Fatalf("loadCache() returned non-empty map for non-existent cache file: %v", cache)
	}

	// mock a cache file with some data
	cache, err = c.refreshCache()
	if err != nil {
		t.Fatalf("error refreshing cache: %v", err)
	}
	mCache := mockCache()
	if len(cache) != len(mCache) {
		t.Fatalf("refreshCache() returned unexpected number of items: %d", len(cache))
	}
	for k, v := range mCache {
		if cache[k] != v {
			t.Errorf("refreshCache() returned unexpected value for key %d: got %q, want %q", k, cache[k], v)
		}
	}

	// check that cache is now populated, after loading
	cache, err = c.loadCache()
	if err != nil {
		t.Fatalf("error loading cache, after saving: %v", err)
	}
	if len(cache) != len(mCache) {
		t.Fatalf("loadCache() returned unexpected number of items: %d", len(cache))
	}
	for k, v := range mCache {
		if cache[k] != v {
			t.Errorf("loadCache() returned unexpected value for key %d: got %q, want %q", k, cache[k], v)
		}
	}
}

func TestFetchProblem_CacheHit(t *testing.T) {
	c := testClient(t)
	// First, refresh the cache to ensure it has the mock problems
	err := c.saveCache(mockCache())
	if err != nil {
		t.Fatalf("error saving cache: %v", err)
	}

	for _, p := range mockProblems {
		if p.IsPaid {
			continue // skip paid problems for this test
		}

		fetchedProblem, err := c.FetchProblem(p.Number)
		if err != nil {
			t.Fatalf("FetchProblem(%d) unexpected error: %v", p.Number, err)
		}
		if !fetchedProblem.IsEqual(p) {
			t.Errorf("FetchProblem(%d) = %+v, want %+v", p.Number, fetchedProblem, p)
		}
	}
}

func TestFetchProblem_CacheMiss(t *testing.T) {
	c := testClient(t)
	// cache is not initialized, so FetchProblem should trigger a cache refresh
	for _, p := range mockProblems {
		if p.IsPaid {
			continue // skip paid problems for this test
		}

		fetchedProblem, err := c.FetchProblem(p.Number)
		if err != nil {
			t.Fatalf("FetchProblem(%d) unexpected error: %v", p.Number, err)
		}
		if !fetchedProblem.IsEqual(p) {
			t.Errorf("FetchProblem(%d) = %+v, want %+v", p.Number, fetchedProblem, p)
		}
	}
}

func TestFetchProblem_Invalid(t *testing.T) {
	c := testClient(t)
	invalidNumbers := []int{999, -1, 0}
	for _, num := range invalidNumbers {
		_, err := c.FetchProblem(num)
		if err == nil {
			t.Errorf("FetchProblem(%d) expected error, got nil", num)
		}
	}
}

func TestFetchProblem_Premium(t *testing.T) {
	c := testClient(t)
	_, err := c.FetchProblem(paidProblem.Number)
	if err == nil {
		t.Errorf("FetchProblem(%d) expected error for premium problem, got nil", paidProblem.Number)
	}
}

func TestFetchDailyProblem(t *testing.T) {
	c := testClient(t)
	fetchedProblem, err := c.FetchDailyProblem()
	if err != nil {
		t.Fatalf("FetchDailyProblem() unexpected error: %v", err)
	}
	if !fetchedProblem.IsEqual(dailyProblem) {
		t.Errorf("FetchDailyProblem() = %+v, want %+v", fetchedProblem, dailyProblem)
	}
}

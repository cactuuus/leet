package leetcode

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// leetcodeURL is the base URL for LeetCode's API.
const leetcodeURL = "https://leetcode.com"

// Client is a client for interacting with the LeetCode API.
type Client struct {
	cachePath 	string
	baseURL 	string
	httpClient 	*http.Client
}

// newClient creates a new LeetCode client with the provided cache path and URL.
// Internal use only, defined here to allow testing with custom values.
func newClient(cachePath string, baseURL string, httpClient *http.Client) (*Client, error) {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &Client{
		cachePath:   cachePath,
		baseURL: 	 baseURL,
		httpClient:  httpClient,
	}, nil
}

// NewClient creates a new LeetCode client with the default cache path and URL.
// This is the way the app is intended to be used. NewClient is instead exposed for testing purposes.
func NewClient() (*Client, error) {
	cachePath, err := defaultCachePath()
	if err != nil {
		return nil, fmt.Errorf("failed to create default cache path: %w", err)
	}
	return newClient(cachePath, leetcodeURL, http.DefaultClient)
}

// defaultCachePath returns the default path for the LeetCode cache file.
// This unfortunately cannot be a constant because it depends on the user's home directory.
func defaultCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cache", "leet", "problems.json"), nil
}

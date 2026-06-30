package leetcode

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/cactuuus/leet/internal/auth"
	"github.com/cactuuus/leet/internal/leetcode/cache"
)

// Client is a client for interacting with the LeetCode API.
type Client struct {
	baseURL 	string
	httpClient 	*http.Client
	credentials auth.Credentials
	cache		*cache.Cache
}

// NewClient creates a new LeetCode client with the desired configuration.
func NewClient(
	cachePath string,
	baseURL string,
	httpClient *http.Client,
	creds auth.Credentials) (*Client, error) {
	return &Client{
		baseURL: 	 baseURL,
		httpClient:  httpClient,
		credentials: creds,
		cache:   	 cache.NewCache(cachePath),
	}, nil
}

// do executes the HTTP request, adding authentication headers if credentials are set.
func(c *Client) do(req *http.Request, referer string) (*http.Response, error) {
	if c.credentials.IsSet() {
		req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.credentials.SessionToken})
		req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.credentials.CSRFToken})
		req.Header.Set("X-CSRFToken", c.credentials.CSRFToken)
	}
	req.Header.Set("Referer", referer)
	return c.httpClient.Do(req)
}

// makeProblemLink constructs the URL for a problem, using the base URL and the problem slug.
func (c *Client) makeProblemLink(slug string) (string, error) {
	return c.makeURL("problems", slug)
}

// makeURL constructs a URL by joining the base URL with the provided path segments.
// It ensures that the resulting URL ends with a single trailing slash, requied for most LeetCode
// endpoints.
func (c *Client) makeURL(segments ...string) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, segments...)
	if err != nil {
		return "", err
	}
	// Ensure the URL ends with a single trailing slash
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	return endpoint, nil
}

// ClearCache clears the cache in memory and saves the empty state to disk.
func (c *Client) ClearCache() error {
	c.cache.Clear()
	return c.cache.Save()
}

// CacheSummary returns a summary of the cache.
func (c *Client) CacheSummary() (string, error) {
	return c.cache.Summary()
}

// CachePath returns the path to the cache file.
func (c *Client) CachePath() string {
	return c.cache.GetPath()
}

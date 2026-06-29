package leetcode

import (
	"fmt"
	"net/http"

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
func(c *Client) do(req *http.Request) (*http.Response, error) {
	if c.credentials.IsSet() {
		req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.credentials.SessionToken})
		req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.credentials.CSRFToken})
		req.Header.Set("X-CSRFToken", c.credentials.CSRFToken)
	}
	return c.httpClient.Do(req)
}

// makeProblemLink constructs the URL for a problem, using the base URL and the problem slug.
func (c *Client) makeProblemLink(slug string) string {
	return fmt.Sprintf("%s/problems/%s", c.baseURL, slug)
}

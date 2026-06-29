package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/cactuuus/leet/internal/problem"
)

// Used to determine if the cache is still valid. Increment on breaking changes that require a cache
// refresh.
const cacheVersion = 1

// Internal structure used solely for JSON serialization.
// Used to avoid exposing private fields in the public Cache struct.
type cacheData struct {
	Version  int                     `json:"version"`
	Previews map[int]problem.Preview `json:"previews"`
}

// Cache represents the structure of the cache file on disk.
type Cache struct {
	path	 string
	isLoaded bool
	data	 cacheData
}

// NewCache creates a new Cache instance, though it does not load the cache from disk. It does so lazily when needed.
func NewCache(path string) *Cache {
	return &Cache{ path: path, data: cacheData{ Version: cacheVersion } }
}

// Load loads the cache from disk.
// If the cache file does not exist, it initializes an empty cache.
// If the cache version does not match, it clears the cache and saves the new, empty version to disk.
func (c *Cache) Load() error {
	data, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		// If the cache file does not exist, simply initialize an empty cache.
		c.data.Previews = make(map[int]problem.Preview)
		c.isLoaded = true
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	// check the version of the cache file before fully unmarshalling it.
	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return fmt.Errorf("failed to unmarshal cache header: %w", err)
	}
	if header.Version != c.data.Version {
		// If the cache version does not match, clear the cache and save the new, empty, version.
		c.data.Previews = make(map[int]problem.Preview)
		return c.Save()
	}

	// else unmarshal the full cache data into the Cache struct.
	if err := json.Unmarshal(data, &c.data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	c.isLoaded = true
	return nil
}

// Save saves the cache to disk.
func (c *Cache) Save() error {
	file, err := os.Create(c.path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(c.data); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

// GetPreview retrieves a problem preview from the cache by its number.
// Returns the preview and a boolean indicating if it was found in the cache.
func (c *Cache) GetPreview(number int) (problem.Preview, bool, error) {
	if !c.isLoaded {
		if err := c.Load(); err != nil {
			return problem.Preview{}, false, fmt.Errorf("failed to lazy-load cache: %w", err)
		}
	}
	preview, ok := c.data.Previews[number]
	return preview, ok, nil
}

// UpdatePreviews updates the cache with the provided problem previews and saves the cache to disk.
func (c *Cache) UpdatePreviews(previews ...problem.Preview) error {
	if !c.isLoaded {
		if err := c.Load(); err != nil {
			return fmt.Errorf("failed to lazy-load cache: %w", err)
		}
	}
	for _, preview := range previews {
		c.data.Previews[preview.Number] = preview
	}
	return c.Save()
}

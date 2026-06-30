package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cactuuus/leet/internal/problem"
)

const (
	// Used to determine if the cache is still valid. Increment on breaking changes that require a
	// cache refresh.
	cacheVersion = 2
	// maximum number of full problems to cache. A bit overkill, but should be light enough not to
	// worry about.
	maxFullCacheSize = 100
)

// dailyProblem represents the structure of the daily problem cache entry.
type dailyProblem struct {
	Problem 	problem.Full 	`json:"problem"`
	ValidUntil 	int64			`json:"valid_until"`
}

// fullCacheEntry represents a cached full problem along with its last accessed timestamp.
type fullCacheEntry struct {
	Problem 		problem.Full	`json:"problem"`
	LastAccessed	int64			`json:"last_accessed"`
}

// Internal structure used solely for JSON serialization.
// Used to avoid exposing private fields in the public Cache struct.
type cacheData struct {
	Version  int                     `json:"version"`
	Previews map[int]problem.Preview `json:"previews"`
	Full	 map[int]fullCacheEntry  `json:"full_problems"`
	Daily    *dailyProblem           `json:"daily"`
}

// Cache represents the structure of the cache file on disk.
type Cache struct {
	path	 string
	isLoaded bool
	data	 cacheData
}

// NewCache creates a new Cache instance, though it does not load the cache from disk. It does so lazily when needed.
func NewCache(path string) *Cache {
	c := Cache{ path: path, isLoaded: false }
	c.Clear() // initialize empty cache
	return &c
}

// Clear clears the cache in memory, resetting it to an empty state. It does not save to disk.
func (c *Cache) Clear() {
	c.data.Version = cacheVersion
	c.data.Previews = make(map[int]problem.Preview)
	c.data.Full = make(map[int]fullCacheEntry)
	c.data.Daily = nil
}

// Load loads the cache from disk.
// If the cache file does not exist, it initializes an empty cache.
// If the cache version does not match, it clears the cache and saves the new, empty version to disk.
func (c *Cache) Load() error {
	if c.isLoaded {
		return nil // already loaded
	}
	data, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		// If the cache file does not exist, simply initialize an empty cache.
		c.Clear()
		c.isLoaded = true
		return c.Save()
	}
	if err != nil {
		return fmt.Errorf("Failed to read cache file:\n%w", err)
	}

	// check the version of the cache file before fully unmarshalling it.
	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return fmt.Errorf("Failed to unmarshal cache header:\n%w", err)
	}
	if header.Version != c.data.Version {
		// If the cache version does not match, clear the cache and save the new, empty, version.
		c.Clear()
		c.isLoaded = true
		return c.Save()
	}

	// else unmarshal the full cache data into the Cache struct.
	if err := json.Unmarshal(data, &c.data); err != nil {
		return fmt.Errorf("Failed to unmarshal cache data:\n%w", err)
	}
	c.isLoaded = true
	return nil
}

// Save saves the cache to disk.
func (c *Cache) Save() error {
	file, err := os.Create(c.path)
	if err != nil {
		return fmt.Errorf("Failed to create cache file:\n%w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(c.data); err != nil {
		return fmt.Errorf("Failed to write cache file:\n%w", err)
	}
	return nil
}

// GetDaily retrieves the daily problem from the cache.
// Returns the problem, a boolean indicating if it was found in the cache, and an error if any occurred.
func (c *Cache) GetDaily() (problem.Full, bool, error) {
	if err := c.Load(); err != nil {
		return problem.Full{}, false, err
	}
	// unavailable or expired
	if c.data.Daily == nil || c.data.Daily.ValidUntil < (time.Now().Unix()) {
		return problem.Full{}, false, nil
	}
	return c.data.Daily.Problem, true, nil
}

// UpdateDaily updates the daily problem in the cache and saves it to disk.
func (c *Cache) UpdateDaily(problem problem.Full, validUntil int64) error {
	if err := c.Load(); err != nil {
		return err
	}
	// update daily entry
	c.data.Daily = &dailyProblem{
		Problem:    problem,
		ValidUntil: validUntil,
	}
	// also update preview (in case it was not already cached)
	c.data.Previews[problem.Number] = problem.Preview
	// lastly, update the full problem cache as well, since we have the full problem data.
	// this also saves to disk, so we don't need to call Save() again.
	return c.UpdateFull(problem)
}

// GetPreview retrieves a problem preview from the cache by its number.
// Returns the preview and a boolean indicating if it was found in the cache.
func (c *Cache) GetPreview(number int) (problem.Preview, bool, error) {
	if err := c.Load(); err != nil {
		return problem.Preview{}, false, err
	}
	preview, ok := c.data.Previews[number]
	return preview, ok, nil
}

// UpdatePreviews updates the cache with the provided problem previews and saves the cache to disk.
func (c *Cache) UpdatePreviews(previews ...problem.Preview) error {
	if err := c.Load(); err != nil {
		return err
	}
	for _, preview := range previews {
		c.data.Previews[preview.Number] = preview
	}
	return c.Save()
}

func (c *Cache) GetFull(number int) (problem.Full, bool, error) {
	if err := c.Load(); err != nil {
		return problem.Full{}, false, err
	}
	entry, ok := c.data.Full[number]
	if ok {
		now := time.Now().Unix()
		entry.LastAccessed = now
		c.data.Full[number] = entry
		// We don't save here to avoid unnecessary disk writes.
		// The next time the cache is saved, this will be persisted.
	}
	return entry.Problem, ok, nil
}

func (c *Cache) UpdateFull(full problem.Full) error {
	if err := c.Load(); err != nil {
		return err
	}
	now := time.Now().Unix()
	c.data.Full[full.Number] = fullCacheEntry{
		Problem:   full,
		LastAccessed: now,
	}

	// If the number of full problems exceeds the maximum allowed, remove the oldest entries until
	// the limit is met.
	for len(c.data.Full) > maxFullCacheSize {
		oldestNum, oldestTimestamp := -1, int64(-1)
		for num := range c.data.Full {
			entry := c.data.Full[num]
			if oldestTimestamp == -1 || entry.LastAccessed < oldestTimestamp {
				oldestTimestamp = entry.LastAccessed
				oldestNum = num
			}
		}
		delete(c.data.Full, oldestNum)
	}
	return c.Save()
}

// Summary returns a human-readable summary of the cache.
func (c *Cache) Summary() (string, error) {
	if err := c.Load(); err != nil {
		return "", err
	}
	// Calculate the size of the cache file.
	sizeStr := "N/A"
	if fi, err := os.Stat(c.path); err == nil {
		sizeKB := float64(fi.Size()) / 1024.0
		sizeMb := sizeKB / 1024.0
		if sizeMb >= 1.0 {
			sizeStr = fmt.Sprintf("%.2f MB", sizeMb)
		} else {
			sizeStr = fmt.Sprintf("%.2f KB", sizeKB)
		}
	}
	// Determine the status of the daily problem in the cache.
	dailyStatus := "Not set/Expired"
	if problem, found, _ := c.GetDaily(); found {
		dailyStatus = fmt.Sprintf("Set to #%d", problem.Number)
	}

	return fmt.Sprintf(
		"CACHE SUMMARY\n\n"+
		"Version.......: %d\n"+
		"Cache Path....: %s\n"+
		"File Size.....: %s\n"+
		"Previews:.....: %d cached\n"+
		"Full Problems.: %d cached (%d max)\n"+
		"Daily Problem.: %s\n",
		c.data.Version, c.path, sizeStr, len(c.data.Previews), len(c.data.Full), maxFullCacheSize, dailyStatus,
	), nil
}

// GetPath returns the path to the cache file.
func (c *Cache) GetPath() string {
	return c.path
}

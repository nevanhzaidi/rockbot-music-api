package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rockbot/music-api/internal/cache"
)

const (
	searchCacheTTL = 5 * time.Minute
	songCacheTTL   = 1 * time.Hour
	httpTimeout    = 10 * time.Second
	maxRetries     = 3
	baseRetryDelay = 500 * time.Millisecond
)

type LrcLibSong struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

type LrcLibClient struct {
	httpClient *http.Client
	baseURL    string
	cache      *cache.RedisCache
}

func NewLrcLibClient(baseURL string, cache *cache.RedisCache) *LrcLibClient {
	return &LrcLibClient{
		httpClient: &http.Client{Timeout: httpTimeout},
		baseURL:    baseURL,
		cache:      cache,
	}
}

// doWithRetry executes an HTTP request with automatic retry and exponential backoff.
//
// Why retry? External APIs like LrcLib can experience transient failures — temporary
// network issues, rate limiting, or server-side errors (5xx). Rather than failing
// immediately, we retry with increasing delays to give the upstream service time to recover.
//
// Retry strategy:
//   - Max attempts: 3 (configurable via maxRetries constant)
//   - Backoff: exponential — 500ms, 1s, 2s (doubles each attempt via bit shift)
//   - Retryable conditions: network errors (timeouts, DNS failures) and 5xx status codes
//   - Non-retryable: 4xx client errors are returned immediately (bad request, not found, etc.)
//
// Important implementation details:
//   - The response body is closed before each retry to prevent resource leaks
//   - On success, the caller is responsible for closing the response body
//   - The last error is wrapped in the final error message for debugging context
//
// Limitations:
//   - Uses time.Sleep which blocks the goroutine; for high-throughput services,
//     a context-based approach with timers would be more appropriate
//   - No jitter is added to the backoff, which could cause thundering herd issues
//     if many clients retry simultaneously
func (c *LrcLibClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Wait before retrying (skip delay on the first attempt)
		if attempt > 0 {
			delay := baseRetryDelay * (1 << (attempt - 1)) // exponential: 500ms, 1s, 2s
			fmt.Printf("retry attempt %d/%d after %v\n", attempt, maxRetries-1, delay)
			time.Sleep(delay)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network-level error (timeout, DNS, connection refused) — retry
			lastErr = err
			continue
		}

		// Server error (5xx) — the upstream is struggling, worth retrying
		// We must close the body before retrying to avoid leaking connections
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("lrclib api returned status: %d", resp.StatusCode)
			continue
		}

		// Success or client error (4xx) — return as-is, no retry needed
		return resp, nil
	}

	return nil, fmt.Errorf("all %d retries failed: %w", maxRetries, lastErr)
}

func (c *LrcLibClient) SearchSongs(query string) (json.RawMessage, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("lrclib:search:%s", query)

	// Check cache first
	if c.cache != nil {
		cached, err := c.cache.Get(ctx, cacheKey)
		if err == nil {
			fmt.Printf("cache hit: %s\n", cacheKey)
			return json.RawMessage(cached), nil
		}
		if err != redis.Nil {
			fmt.Printf("cache error for key %s: %v\n", cacheKey, err)
		}
	}

	// Cache miss — fetch from LrcLib API
	fmt.Printf("cache miss: %s\n", cacheKey)
	endpoint := fmt.Sprintf("%s/search?q=%s", c.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "RockbotMusicAPI/1.0")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("calling lrclib api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib api returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Store in cache
	if c.cache != nil {
		if setErr := c.cache.Set(ctx, cacheKey, string(body), searchCacheTTL); setErr != nil {
			fmt.Printf("cache set error for key %s: %v\n", cacheKey, setErr)
		}
	}

	return json.RawMessage(body), nil
}

func (c *LrcLibClient) GetSongByID(id int) (*LrcLibSong, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("lrclib:song:%d", id)

	// Check cache first
	if c.cache != nil {
		cached, err := c.cache.Get(ctx, cacheKey)
		if err == nil {
			var song LrcLibSong
			if jsonErr := json.Unmarshal([]byte(cached), &song); jsonErr == nil {
				fmt.Printf("cache hit: %s\n", cacheKey)
				return &song, nil
			}
		}
		if err != redis.Nil {
			fmt.Printf("cache error for key %s: %v\n", cacheKey, err)
		}
	}

	// Cache miss — fetch from LrcLib API
	fmt.Printf("cache miss: %s\n", cacheKey)
	endpoint := fmt.Sprintf("%s/get/%d", c.baseURL, id)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "RockbotMusicAPI/1.0")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("calling lrclib api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib api returned status: %d", resp.StatusCode)
	}

	var song LrcLibSong
	if err := json.NewDecoder(resp.Body).Decode(&song); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Store in cache
	if c.cache != nil {
		songJSON, _ := json.Marshal(song)
		if setErr := c.cache.Set(ctx, cacheKey, string(songJSON), songCacheTTL); setErr != nil {
			fmt.Printf("cache set error for key %s: %v\n", cacheKey, setErr)
		}
	}

	return &song, nil
}

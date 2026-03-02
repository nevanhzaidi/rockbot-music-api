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
		httpClient: &http.Client{},
		baseURL:    baseURL,
		cache:      cache,
	}
}

func (c *LrcLibClient) SearchSongs(query string) (json.RawMessage, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("lrclib:search:%s", query)

	// Check cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil {
		fmt.Printf("cache hit: %s\n", cacheKey)
		return json.RawMessage(cached), nil
	}
	if err != redis.Nil {
		fmt.Printf("cache error for key %s: %v\n", cacheKey, err)
	}

	// Cache miss — fetch from LrcLib API
	fmt.Printf("cache miss: %s\n", cacheKey)
	endpoint := fmt.Sprintf("%s/search?q=%s", c.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "RockbotMusicAPI/1.0")

	resp, err := c.httpClient.Do(req)
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
	if setErr := c.cache.Set(ctx, cacheKey, string(body), searchCacheTTL); setErr != nil {
		fmt.Printf("cache set error for key %s: %v\n", cacheKey, setErr)
	}

	return json.RawMessage(body), nil
}

func (c *LrcLibClient) GetSongByID(id int) (*LrcLibSong, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("lrclib:song:%d", id)

	// Check cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil {
		var song LrcLibSong
		if jsonErr := json.Unmarshal([]byte(cached), &song); jsonErr == nil {
			fmt.Printf("cache hit: %s\n", cacheKey)
			return &song, nil
		}
	}
	if err != nil && err != redis.Nil {
		fmt.Printf("cache error for key %s: %v\n", cacheKey, err)
	}

	// Cache miss — fetch from LrcLib API
	fmt.Printf("cache miss: %s\n", cacheKey)
	endpoint := fmt.Sprintf("%s/get/%d", c.baseURL, id)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "RockbotMusicAPI/1.0")

	resp, err := c.httpClient.Do(req)
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
	songJSON, _ := json.Marshal(song)
	if setErr := c.cache.Set(ctx, cacheKey, string(songJSON), songCacheTTL); setErr != nil {
		fmt.Printf("cache set error for key %s: %v\n", cacheKey, setErr)
	}

	return &song, nil
}

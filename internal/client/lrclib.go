package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// LrcLibClient handles communication with the LrcLib API.
type LrcLibClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewLrcLibClient creates a new LrcLib API client.
// baseURL is the LrcLib API base URL (e.g. "https://lrclib.net/api").
func NewLrcLibClient(baseURL string) *LrcLibClient {
	return &LrcLibClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

// SearchSongs searches for songs by a query string.
// It calls GET https://lrclib.net/api/search?q=<query>
// Returns the raw JSON response from the API.
func (c *LrcLibClient) SearchSongs(query string) (json.RawMessage, error) {
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

	return json.RawMessage(body), nil
}

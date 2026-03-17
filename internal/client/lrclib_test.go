package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newTestClient creates an LrcLibClient pointing at a test server with no cache.
// We pass nil for cache since these tests focus on HTTP behavior and retry logic.
func newTestClient(serverURL string) *LrcLibClient {
	return &LrcLibClient{
		httpClient: &http.Client{},
		baseURL:    serverURL,
		cache:      nil,
	}
}

// --- doWithRetry tests ---

func TestDoWithRetry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.doWithRetry(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got: %d", resp.StatusCode)
	}
}

func TestDoWithRetry_RetriesOn5xx(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.doWithRetry(req)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got: %d", resp.StatusCode)
	}

	if callCount.Load() != 3 {
		t.Errorf("expected 3 attempts, got: %d", callCount.Load())
	}
}

func TestDoWithRetry_AllRetriesFail(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	_, err := client.doWithRetry(req)
	if err == nil {
		t.Fatal("expected error after all retries failed")
	}

	if callCount.Load() != int32(maxRetries) {
		t.Errorf("expected %d attempts, got: %d", maxRetries, callCount.Load())
	}
}

func TestDoWithRetry_NoRetryOn4xx(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.doWithRetry(req)
	if err != nil {
		t.Fatalf("expected no error for 4xx, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got: %d", resp.StatusCode)
	}

	if callCount.Load() != 1 {
		t.Errorf("expected 1 attempt (no retry on 4xx), got: %d", callCount.Load())
	}
}

// --- SearchSongs tests (without cache) ---

func TestSearchSongs_Success(t *testing.T) {
	songs := []LrcLibSong{
		{ID: 1, TrackName: "Bohemian Rhapsody", ArtistName: "Queen"},
	}
	songsJSON, _ := json.Marshal(songs)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("expected path /search, got: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "queen" {
			t.Errorf("expected query param q=queen, got: %s", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(songsJSON)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.SearchSongs("queen")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var parsed []LrcLibSong
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 song, got: %d", len(parsed))
	}
	if parsed[0].TrackName != "Bohemian Rhapsody" {
		t.Errorf("expected track name 'Bohemian Rhapsody', got: %s", parsed[0].TrackName)
	}
}

func TestSearchSongs_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.SearchSongs("test")
	if err == nil {
		t.Fatal("expected error on bad request")
	}
}

// --- GetSongByID tests (without cache) ---

func TestGetSongByID_Success(t *testing.T) {
	song := LrcLibSong{
		ID:         3396,
		TrackName:  "Bohemian Rhapsody",
		ArtistName: "Queen",
		AlbumName:  "A Night at the Opera",
		Duration:   354.32,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/3396" {
			t.Errorf("expected path /get/3396, got: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(song)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetSongByID(3396)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.TrackName != "Bohemian Rhapsody" {
		t.Errorf("expected 'Bohemian Rhapsody', got: %s", result.TrackName)
	}
	if result.ArtistName != "Queen" {
		t.Errorf("expected 'Queen', got: %s", result.ArtistName)
	}
}

func TestGetSongByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetSongByID(99999)
	if err == nil {
		t.Fatal("expected error on not found")
	}
}

func TestGetSongByID_UserAgentHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "RockbotMusicAPI/1.0" {
			t.Errorf("expected User-Agent 'RockbotMusicAPI/1.0', got: %s", ua)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LrcLibSong{ID: 1})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.GetSongByID(1)
}

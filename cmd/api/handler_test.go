package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got: %s", body["status"])
	}
}

func TestSearchHandler_MissingQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/songs/search", nil)
	w := httptest.NewRecorder()

	handler := searchHandler(nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)

	if body["error"] != "query parameter 'q' is required" {
		t.Errorf("unexpected error message: %s", body["error"])
	}
}

func TestCreatePlaylistHandler_MissingName(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/playlists", strings.NewReader(`{"description":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := createPlaylistHandler(nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)

	if body["error"] != "name is required" {
		t.Errorf("unexpected error message: %s", body["error"])
	}
}

func TestCreatePlaylistHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/playlists", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := createPlaylistHandler(nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)

	if body["error"] != "invalid request body" {
		t.Errorf("unexpected error message: %s", body["error"])
	}
}

func TestAddSongHandler_MissingLrclibID(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/playlists/123/songs", strings.NewReader(`{"rating":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := addSongHandler(nil, nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)

	if body["error"] != "lrclib_id is required" {
		t.Errorf("unexpected error message: %s", body["error"])
	}
}

func TestAddSongHandler_InvalidRating(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		errMsg string
	}{
		{"rating too high", `{"lrclib_id":1,"rating":6}`, "rating must be between 0 and 5"},
		{"rating negative", `{"lrclib_id":1,"rating":-1}`, "rating must be between 0 and 5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/playlists/123/songs", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := addSongHandler(nil, nil)
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got: %d", w.Code)
			}

			var body map[string]string
			json.NewDecoder(w.Body).Decode(&body)

			if body["error"] != tt.errMsg {
				t.Errorf("expected error '%s', got: '%s'", tt.errMsg, body["error"])
			}
		})
	}
}

func TestAddSongHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/playlists/123/songs", strings.NewReader(`bad json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := addSongHandler(nil, nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got: %s", contentType)
	}
}

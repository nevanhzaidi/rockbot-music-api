package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rockbot/music-api/internal/cache"
	"github.com/rockbot/music-api/internal/client"
	"github.com/rockbot/music-api/internal/repository"
)

func main() {
	// Load .env file if present (optional in Docker where env vars are injected)
	godotenv.Load()

	lrclibURL := os.Getenv("LRCLIB_BASE_URL")
	if lrclibURL == "" {
		log.Fatal("LRCLIB_BASE_URL environment variable is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL")

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisCache := cache.NewRedisCache(redisAddr)
	if err := redisCache.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis")

	lrclib := client.NewLrcLibClient(lrclibURL, redisCache)
	playlistRepo := repository.NewPlaylistRepository(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/songs/search", searchHandler(lrclib))
	mux.HandleFunc("POST /api/playlists", createPlaylistHandler(playlistRepo))
	mux.HandleFunc("GET /api/playlists/{id}", getPlaylistHandler(playlistRepo))
	mux.HandleFunc("POST /api/playlists/{id}/songs", addSongHandler(playlistRepo, lrclib))
	mux.HandleFunc("DELETE /api/playlists/{id}/songs/{songId}", deleteSongHandler(playlistRepo))

	fmt.Println("Starting Rockbot Music API server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func searchHandler(lrclib *client.LrcLibClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "query parameter 'q' is required"})
			return
		}

		songs, err := lrclib.SearchSongs(query)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(songs)
	}
}

func createPlaylistHandler(repo *repository.PlaylistRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if input.Name == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
			return
		}

		playlist, err := repo.CreatePlaylist(input.Name, input.Description)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(playlist)
	}
}

func getPlaylistHandler(repo *repository.PlaylistRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		playlist, songs, err := repo.GetPlaylist(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "playlist not found"})
			return
		}

		response := map[string]any{
			"playlist": playlist,
			"songs":    songs,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func addSongHandler(repo *repository.PlaylistRepository, lrclib *client.LrcLibClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistID := r.PathValue("id")

		var input struct {
			LrclibID int    `json:"lrclib_id"`
			Rating   int    `json:"rating"`
			Notes    string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if input.LrclibID == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "lrclib_id is required"})
			return
		}

		if input.Rating < 0 || input.Rating > 5 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "rating must be between 0 and 5"})
			return
		}

		apiSong, err := lrclib.GetSongByID(input.LrclibID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch song from LrcLib"})
			return
		}

		song, err := repo.AddSong(playlistID, apiSong.TrackName, apiSong.ArtistName, apiSong.AlbumName, apiSong.Duration, input.Rating, input.Notes)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(song)
	}
}

func deleteSongHandler(repo *repository.PlaylistRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistID := r.PathValue("id")
		songID := r.PathValue("songId")

		if err := repo.DeleteSong(playlistID, songID); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "song not found"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "song deleted"})
	}
}

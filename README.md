# Rockbot Music API

A Go backend service that integrates with the [LrcLib](https://lrclib.net/docs) music API, allowing users to search songs, create playlists, and save tracks with personal ratings and notes. Backed by PostgreSQL for persistent storage and Redis for caching.

## Table of Contents

- [How to Run](#how-to-run)
- [API Endpoints](#api-endpoints)
- [Example Requests](#example-requests)
- [Project Structure](#project-structure)
- [Architecture Decisions](#architecture-decisions)
- [Assumptions & Trade-offs](#assumptions--trade-offs)
- [Testing](#testing)
- [AI Usage](#ai-usage)

## How to Run

### Option 1: Docker Compose (Recommended)

Spins up the Go API, PostgreSQL, and Redis with a single command. No local setup required.

```bash
docker-compose up --build
```

The API will be available at `http://localhost:8080`. The database migration is applied automatically on first startup.

To stop:
```bash
docker-compose down
```

To reset the database:
```bash
docker-compose down -v
docker-compose up --build
```

### Option 2: Run Natively

Prerequisites:
- Go 1.23+
- PostgreSQL running locally
- Redis running locally

```bash
# 1. Clone and enter the project
cd rockbot-music-api

# 2. Create your .env file
cp .env.example .env

# 3. Create the database and apply migration
psql -U postgres -c "CREATE DATABASE rockbot;"
psql -U postgres -d rockbot -f migrations/001_create_tables.sql

# 4. Run the server
go run ./cmd/api/
```

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `LRCLIB_BASE_URL` | LrcLib API base URL | `https://lrclib.net/api` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/rockbot?sslmode=disable` |
| `REDIS_URL` | Redis address | `localhost:6379` |

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/songs/search?q=<query>` | Search songs via LrcLib |
| `POST` | `/api/playlists` | Create a new playlist |
| `GET` | `/api/playlists/{id}` | Get playlist with all songs |
| `POST` | `/api/playlists/{id}/songs` | Add a song to a playlist |
| `DELETE` | `/api/playlists/{id}/songs/{songId}` | Remove a song from a playlist |

## Example Requests

### Health Check
```bash
curl http://localhost:8080/health
```
```json
{"status": "ok"}
```

### Search Songs
```bash
curl "http://localhost:8080/api/songs/search?q=bohemian+rhapsody"
```
```json
[
  {
    "id": 3396,
    "trackName": "Bohemian Rhapsody",
    "artistName": "Queen",
    "albumName": "A Night at the Opera",
    "duration": 354.32,
    "instrumental": false,
    "plainLyrics": "...",
    "syncedLyrics": "..."
  }
]
```

### Create a Playlist
```bash
curl -X POST http://localhost:8080/api/playlists \
  -H "Content-Type: application/json" \
  -d '{"name": "Rock Classics", "description": "Best rock songs ever"}'
```
```json
{
  "id": "a1b2c3d4-...",
  "name": "Rock Classics",
  "description": "Best rock songs ever",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Add a Song to a Playlist

Use the `id` from search results as `lrclib_id`. The API fetches the song metadata from LrcLib automatically.

```bash
curl -X POST http://localhost:8080/api/playlists/<playlist-id>/songs \
  -H "Content-Type: application/json" \
  -d '{"lrclib_id": 3396, "rating": 5, "notes": "Freddie Mercury classic"}'
```
```json
{
  "id": "e5f6g7h8-...",
  "playlist_id": "a1b2c3d4-...",
  "track_name": "Bohemian Rhapsody",
  "artist_name": "Queen",
  "album_name": "A Night at the Opera",
  "duration": 354.32,
  "rating": 5,
  "notes": "Freddie Mercury classic",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Get a Playlist
```bash
curl http://localhost:8080/api/playlists/<playlist-id>
```
```json
{
  "playlist": {
    "id": "a1b2c3d4-...",
    "name": "Rock Classics",
    "description": "Best rock songs ever",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "songs": [
    {
      "id": "e5f6g7h8-...",
      "track_name": "Bohemian Rhapsody",
      "artist_name": "Queen",
      "rating": 5,
      "notes": "Freddie Mercury classic"
    }
  ]
}
```

### Delete a Song from a Playlist
```bash
curl -X DELETE http://localhost:8080/api/playlists/<playlist-id>/songs/<song-id>
```
```json
{"message": "song deleted"}
```

## Project Structure

```
rockbot-music-api/
├── cmd/
│   └── api/
│       ├── main.go              # Application entrypoint, routes, handlers
│       └── handler_test.go      # Handler unit tests
├── internal/
│   ├── cache/
│   │   └── redis.go             # Redis cache wrapper (Get/Set with TTL)
│   ├── client/
│   │   ├── lrclib.go            # LrcLib API client with caching and retry
│   │   └── lrclib_test.go       # Client unit tests (retry, search, get by ID)
│   ├── model/
│   │   └── playlist.go          # Data models (Playlist, PlaylistSong)
│   └── repository/
│       └── playlist.go          # PostgreSQL repository (CRUD operations)
├── migrations/
│   └── 001_create_tables.sql    # Database schema
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Full stack: API + PostgreSQL + Redis
├── .env.example                 # Environment variable template
├── go.mod
└── go.sum
```

## Architecture Decisions

### Why LrcLib?
LrcLib is a free, no-auth music metadata and lyrics API. It provides song search, track metadata, and lyrics — ideal for building a playlist service without API key management.

### Why PostgreSQL?
Playlists and songs are relational data with foreign key constraints (songs belong to playlists). PostgreSQL provides UUID generation, CHECK constraints for rating validation, and CASCADE deletes — all used in the schema.

### Why Redis for Caching?
LrcLib API responses are cached in Redis to reduce external API calls and improve response times. Two TTL strategies are used:
- **Search results**: 5-minute TTL (results may change as new songs are added)
- **Song by ID**: 1-hour TTL (individual song metadata is stable)

Cache errors are handled gracefully — if Redis is unavailable, the API falls back to calling LrcLib directly.

### Retry with Exponential Backoff
The LrcLib client retries failed requests up to 3 times with exponential backoff (500ms, 1s, 2s). Only network errors and 5xx responses trigger retries. Client errors (4xx) are returned immediately. See `doWithRetry` in `internal/client/lrclib.go` for the thoroughly documented implementation.

### HTTP Client Timeout
A 10-second timeout is set on the HTTP client to prevent requests from hanging indefinitely if LrcLib is unresponsive.

## Assumptions & Trade-offs

- **No authentication**: The API is open. In production, I would add JWT or API key authentication middleware.
- **No pagination**: Playlist songs and search results are returned in full. For large datasets, cursor-based pagination would be appropriate.
- **Rating 0-5**: Validated at both the API layer and database level (CHECK constraint). A rating of 0 means unrated.
- **Search proxies raw JSON**: The search endpoint returns LrcLib's raw response. This avoids unnecessary serialization/deserialization but means the response format is coupled to LrcLib's API.
- **Single migration file**: For a project of this scope, a single SQL file is sufficient. For production, I would use a migration tool like golang-migrate with versioned up/down migrations.
- **Repository tests**: The repository layer is tested indirectly through handler tests for validation logic. Full repository tests would be integration tests requiring a live database, which is a deliberate testing boundary decision.
- **No graceful shutdown**: The server uses `log.Fatal(http.ListenAndServe(...))`. In production, I would implement signal handling with `context` for graceful shutdown of HTTP server, database, and Redis connections.

## Testing

Run all tests (no external dependencies required):
```bash
go test ./...
```

Run with verbose output:
```bash
go test ./... -v
```

### Test Coverage

- **Client tests** (9 tests): retry logic (success, 5xx retry, all retries fail, no retry on 4xx), search songs, get song by ID, User-Agent header
- **Handler tests** (8 tests): health check, missing query params, missing playlist name, invalid JSON, missing lrclib_id, invalid rating (too high, negative), content-type validation

Tests use `httptest.NewServer` to mock the LrcLib API — no real network calls, Redis, or PostgreSQL needed.

## AI Usage

AI tools (Claude) were used to assist with boilerplate generation and code scaffolding. All architectural decisions, API design choices, and implementation logic were my own. Specifically:

- **Where used**: Generating initial file structures, boilerplate handler patterns, Dockerfile template, and test scaffolding
- **How it affected workflow**: Sped up repetitive tasks (writing similar handler functions, test setup patterns) so I could focus more time on design decisions and reliability features
- **Trade-offs**: AI-generated code was reviewed and modified to match the project's patterns and conventions. The risk of accepting generated code without review was mitigated by running tests and manual verification after each change

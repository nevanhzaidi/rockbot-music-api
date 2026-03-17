package model

import "time"

type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlaylistSong struct {
	ID         string    `json:"id"`
	PlaylistID string    `json:"playlist_id"`
	LrclibID   int       `json:"lrclib_id,omitempty"` // LrcLib track ID; used to fetch lyrics
	TrackName  string    `json:"track_name"`
	ArtistName string    `json:"artist_name"`
	AlbumName  string    `json:"album_name"`
	Duration   float64   `json:"duration"`
	Rating     int       `json:"rating"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

package repository

import (
	"database/sql"
	"fmt"

	"github.com/rockbot/music-api/internal/model"
)

type PlaylistRepository struct {
	db *sql.DB
}

func NewPlaylistRepository(db *sql.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) CreatePlaylist(name, description string) (*model.Playlist, error) {
	var p model.Playlist
	err := r.db.QueryRow(
		`INSERT INTO playlists (name, description) VALUES ($1, $2)
		 RETURNING id, name, description, created_at, updated_at`,
		name, description,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating playlist: %w", err)
	}
	return &p, nil
}

func (r *PlaylistRepository) GetPlaylist(id string) (*model.Playlist, []model.PlaylistSong, error) {
	var p model.Playlist
	err := r.db.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM playlists WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("getting playlist: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT id, playlist_id, track_name, artist_name, album_name, duration, rating, notes, created_at, updated_at
		 FROM playlist_songs WHERE playlist_id = $1 ORDER BY created_at DESC`,
		id,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("getting playlist songs: %w", err)
	}
	defer rows.Close()

	var songs []model.PlaylistSong
	for rows.Next() {
		var s model.PlaylistSong
		err := rows.Scan(&s.ID, &s.PlaylistID, &s.TrackName, &s.ArtistName, &s.AlbumName,
			&s.Duration, &s.Rating, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("scanning playlist song: %w", err)
		}
		songs = append(songs, s)
	}

	return &p, songs, nil
}

func (r *PlaylistRepository) AddSong(playlistID, trackName, artistName, albumName string, duration float64, rating int, notes string) (*model.PlaylistSong, error) {
	var s model.PlaylistSong
	err := r.db.QueryRow(
		`INSERT INTO playlist_songs (playlist_id, track_name, artist_name, album_name, duration, rating, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, playlist_id, track_name, artist_name, album_name, duration, rating, notes, created_at, updated_at`,
		playlistID, trackName, artistName, albumName, duration, rating, notes,
	).Scan(&s.ID, &s.PlaylistID, &s.TrackName, &s.ArtistName, &s.AlbumName,
		&s.Duration, &s.Rating, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding song: %w", err)
	}
	return &s, nil
}

func (r *PlaylistRepository) DeleteSong(playlistID, songID string) error {
	result, err := r.db.Exec(
		`DELETE FROM playlist_songs WHERE id = $1 AND playlist_id = $2`,
		songID, playlistID,
	)
	if err != nil {
		return fmt.Errorf("deleting song: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("song not found")
	}

	return nil
}

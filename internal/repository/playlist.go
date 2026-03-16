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
		`SELECT id, playlist_id, lrclib_id, track_name, artist_name, album_name, duration, rating, notes, created_at, updated_at
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
		var lrclibID sql.NullInt64
		err := rows.Scan(&s.ID, &s.PlaylistID, &lrclibID, &s.TrackName, &s.ArtistName, &s.AlbumName,
			&s.Duration, &s.Rating, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("scanning playlist song: %w", err)
		}
		if lrclibID.Valid {
			s.LrclibID = int(lrclibID.Int64)
		}
		songs = append(songs, s)
	}

	return &p, songs, nil
}

func (r *PlaylistRepository) ListPlaylists() ([]model.Playlist, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, created_at, updated_at FROM playlists ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing playlists: %w", err)
	}
	defer rows.Close()

	var list []model.Playlist
	for rows.Next() {
		var p model.Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning playlist: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

// SongExistsInPlaylist returns true if the playlist already contains a track with the given lrclib_id.
func (r *PlaylistRepository) SongExistsInPlaylist(playlistID string, lrclibID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM playlist_songs WHERE playlist_id = $1 AND lrclib_id = $2)`,
		playlistID, lrclibID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking duplicate: %w", err)
	}
	return exists, nil
}

func (r *PlaylistRepository) AddSong(playlistID string, lrclibID int, trackName, artistName, albumName string, duration float64, rating int, notes string) (*model.PlaylistSong, error) {
	var s model.PlaylistSong
	err := r.db.QueryRow(
		`INSERT INTO playlist_songs (playlist_id, lrclib_id, track_name, artist_name, album_name, duration, rating, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, playlist_id, lrclib_id, track_name, artist_name, album_name, duration, rating, notes, created_at, updated_at`,
		playlistID, lrclibID, trackName, artistName, albumName, duration, rating, notes,
	).Scan(&s.ID, &s.PlaylistID, &s.LrclibID, &s.TrackName, &s.ArtistName, &s.AlbumName,
		&s.Duration, &s.Rating, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding song: %w", err)
	}
	return &s, nil
}

func (r *PlaylistRepository) DeletePlaylist(id string) error {
	result, err := r.db.Exec(`DELETE FROM playlists WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting playlist: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("playlist not found")
	}
	return nil
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

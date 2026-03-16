-- Add lrclib_id to playlist_songs so we can fetch lyrics for any song in a playlist.
-- Existing rows will have NULL; new adds will populate it.
ALTER TABLE playlist_songs
ADD COLUMN IF NOT EXISTS lrclib_id INT DEFAULT NULL;

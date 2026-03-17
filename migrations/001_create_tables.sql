CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS playlists (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlist_songs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_name  VARCHAR(255) NOT NULL,
    artist_name VARCHAR(255) NOT NULL,
    album_name  VARCHAR(255) DEFAULT '',
    duration    FLOAT DEFAULT 0,
    rating      INT DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
    notes       TEXT DEFAULT '',
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

import { useState, useEffect } from 'react';
import { getLyrics } from '../api';
import './LyricsModal.css';

export default function LyricsModal({ lrclibId, trackName, artistName, onClose }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!lrclibId) return;
    setLoading(true);
    setError(null);
    getLyrics(lrclibId)
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [lrclibId]);

  const lyrics = data?.plain_lyrics || '';
  const displayTitle = trackName || data?.track_name || 'Lyrics';
  const displayArtist = artistName || data?.artist_name || '';

  return (
    <div className="lyrics-modal-overlay" onClick={onClose}>
      <div className="lyrics-modal" onClick={(e) => e.stopPropagation()}>
        <div className="lyrics-modal-header">
          <h2>{displayTitle}</h2>
          {displayArtist && <p className="lyrics-artist">{displayArtist}</p>}
          <button type="button" className="lyrics-close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>
        <div className="lyrics-modal-body">
          {loading && <p className="lyrics-loading">Loading lyrics…</p>}
          {error && <p className="lyrics-error">{error}</p>}
          {!loading && !error && (
            <pre className="lyrics-text">{lyrics || 'No lyrics available.'}</pre>
          )}
        </div>
      </div>
    </div>
  );
}

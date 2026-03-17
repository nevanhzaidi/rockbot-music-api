import { useState } from 'react';
import { searchSongs, addSongToPlaylist } from '../api';
import './AddSongModal.css';

export default function AddSongModal({ playlistId, onAdded, onClose }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [searching, setSearching] = useState(false);
  const [rating, setRating] = useState(0);
  const [notes, setNotes] = useState('');
  const [selected, setSelected] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!query.trim()) return;
    setSearching(true);
    setError(null);
    setResults([]);
    try {
      const data = await searchSongs(query.trim());
      setResults(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message);
    } finally {
      setSearching(false);
    }
  };

  const handleAdd = async () => {
    if (!selected) return;
    setSubmitting(true);
    setError(null);
    try {
      await addSongToPlaylist(playlistId, selected.id, rating, notes);
      onAdded?.();
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="add-song-overlay" onClick={onClose}>
      <div className="add-song-modal" onClick={(e) => e.stopPropagation()}>
        <div className="add-song-header">
          <h2>Add song to playlist</h2>
          <button type="button" className="add-song-close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>
        <form onSubmit={handleSearch} className="add-song-search">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search by track or artist…"
            autoFocus
          />
          <button type="submit" disabled={searching}>
            {searching ? 'Searching…' : 'Search'}
          </button>
        </form>
        {error && <p className="add-song-error">{error}</p>}
        <div className="add-song-results">
          {results.length > 0 && !selected && (
            <ul>
              {results.slice(0, 8).map((s) => (
                <li key={s.id}>
                  <button
                    type="button"
                    className="add-song-result-item"
                    onClick={() => setSelected(s)}
                  >
                    <span className="add-song-result-title">{s.trackName}</span>
                    <span className="add-song-result-artist">{s.artistName}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
          {selected && (
            <div className="add-song-selected">
              <p className="add-song-selected-title">
                <strong>{selected.trackName}</strong> — {selected.artistName}
              </p>
              <label>
                Rating (0–5)
                <input
                  type="number"
                  min={0}
                  max={5}
                  value={rating}
                  onChange={(e) => setRating(Number(e.target.value))}
                />
              </label>
              <label>
                Notes
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="Optional notes…"
                  rows={2}
                />
              </label>
              <div className="add-song-actions">
                <button type="button" onClick={() => setSelected(null)}>
                  Back
                </button>
                <button type="button" onClick={handleAdd} disabled={submitting}>
                  {submitting ? 'Adding…' : 'Add to playlist'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

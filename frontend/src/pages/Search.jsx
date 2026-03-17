import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { searchSongs } from '../api';
import AddSongModal from '../components/AddSongModal';
import './Search.css';

export default function Search() {
  const [searchParams] = useSearchParams();
  const addToPlaylistId = searchParams.get('addTo');
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [searching, setSearching] = useState(false);
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

  return (
    <div className="page search-page">
      <h1>Search songs</h1>
      <form onSubmit={handleSearch} className="search-form">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Track or artist…"
          autoFocus
        />
        <button type="submit" disabled={searching}>
          {searching ? 'Searching…' : 'Search'}
        </button>
      </form>
      {error && <p className="page-error">{error}</p>}
      {!searching && results.length === 0 && query.trim() !== '' && !error && (
        <div className="search-empty">No results for &quot;{query}&quot;. Try another search.</div>
      )}
      {results.length > 0 && (
        <ul className="search-results">
          {results.map((s) => (
            <li key={s.id} className="search-result-item">
              <div className="search-result-info">
                <span className="search-result-title">{s.trackName}</span>
                <span className="search-result-meta">
                  {s.artistName}
                  {s.albumName && ` · ${s.albumName}`}
                </span>
              </div>
              <div className="search-result-actions">
                <Link
                  to={addToPlaylistId ? `/playlists/${addToPlaylistId}` : '/playlists'}
                  className="btn btn-primary"
                >
                  {addToPlaylistId ? 'Add to playlist' : 'Pick a playlist'}
                </Link>
              </div>
            </li>
          ))}
        </ul>
      )}
      {results.length > 0 && !addToPlaylistId && (
        <p className="search-hint">
          Open a playlist and use &quot;Add song&quot; to add search results to it.
        </p>
      )}
    </div>
  );
}

import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { listPlaylists, createPlaylist } from '../api';
import './Playlists.css';

export default function Playlists() {
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [creating, setCreating] = useState(false);

  const load = () => {
    setLoading(true);
    setError(null);
    listPlaylists()
      .then(setList)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => load(), []);

  const handleCreate = async (e) => {
    e.preventDefault();
    if (!name.trim()) return;
    setCreating(true);
    setError(null);
    try {
      await createPlaylist(name.trim(), description.trim());
      setName('');
      setDescription('');
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="page playlists-page">
      <h1>Playlists</h1>
      <form onSubmit={handleCreate} className="playlists-create">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Playlist name"
          required
        />
        <input
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Description (optional)"
        />
        <button type="submit" disabled={creating}>
          {creating ? 'Creating…' : 'Create playlist'}
        </button>
      </form>
      {error && <p className="page-error">{error}</p>}
      {loading && <p className="page-muted">Loading playlists…</p>}
      {!loading && list.length === 0 && (
        <div className="playlists-empty">No playlists yet. Create one using the form above.</div>
      )}
      {!loading && list.length > 0 && (
        <ul className="playlists-list">
          {list.map((p) => (
            <li key={p.id}>
              <Link to={`/playlists/${p.id}`} className="playlist-card">
                <span className="playlist-card-name">{p.name}</span>
                {p.description && (
                  <span className="playlist-card-desc">{p.description}</span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

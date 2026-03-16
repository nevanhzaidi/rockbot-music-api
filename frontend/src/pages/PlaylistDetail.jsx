import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getPlaylist, removeSongFromPlaylist, deletePlaylist } from '../api';
import LyricsModal from '../components/LyricsModal';
import AddSongModal from '../components/AddSongModal';
import './PlaylistDetail.css';

export default function PlaylistDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [playlist, setPlaylist] = useState(null);
  const [songs, setSongs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lyricsFor, setLyricsFor] = useState(null);
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [removingId, setRemovingId] = useState(null);
  const [deleting, setDeleting] = useState(false);

  const load = () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    getPlaylist(id)
      .then((data) => {
        setPlaylist(data.playlist);
        setSongs(data.songs || []);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => load(), [id]);

  const handleRemove = async (songId) => {
    setRemovingId(songId);
    try {
      await removeSongFromPlaylist(id, songId);
      setSongs((prev) => prev.filter((s) => s.id !== songId));
    } catch (e) {
      setError(e.message);
    } finally {
      setRemovingId(null);
    }
  };

  const handleDeletePlaylist = async () => {
    if (!window.confirm(`Delete "${playlist.name}"? This cannot be undone.`)) return;
    setDeleting(true);
    setError(null);
    try {
      await deletePlaylist(id);
      navigate('/playlists');
    } catch (e) {
      setError(e.message);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) return <div className="page"><p className="page-muted">Loading playlist…</p></div>;
  if (error && !playlist) return <div className="page"><p className="page-error">{error}</p></div>;
  if (!playlist) return null;

  return (
    <div className="page playlist-detail-page">
      <div className="playlist-detail-header">
        <Link to="/playlists" className="playlist-back">← Playlists</Link>
        <h1>{playlist.name}</h1>
        {playlist.description && (
          <p className="playlist-detail-desc">{playlist.description}</p>
        )}
        <div className="playlist-detail-actions">
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => setAddModalOpen(true)}
          >
            Add song
          </button>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleDeletePlaylist}
            disabled={deleting}
          >
            {deleting ? 'Deleting…' : 'Delete playlist'}
          </button>
        </div>
      </div>
      {error && <p className="page-error">{error}</p>}
      {songs.length === 0 ? (
        <div className="playlist-detail-empty">No songs yet. Click &quot;Add song&quot; above to search and add tracks.</div>
      ) : (
        <ul className="playlist-songs">
          {songs.map((s) => (
            <li key={s.id} className="playlist-song-item">
              <div className="playlist-song-info">
                <span className="playlist-song-title">{s.track_name}</span>
                <span className="playlist-song-meta">
                  {s.artist_name}
                  {s.rating > 0 && ` · ★ ${s.rating}`}
                  {s.notes && ` · ${s.notes}`}
                </span>
              </div>
              <div className="playlist-song-actions">
                {s.lrclib_id ? (
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={() => setLyricsFor({ id: s.lrclib_id, trackName: s.track_name, artistName: s.artist_name })}
                  >
                    View lyrics
                  </button>
                ) : (
                  <span className="playlist-song-no-lyrics" title="Lyrics only for songs added after the lyrics feature">—</span>
                )}
                <button
                  type="button"
                  className="btn btn-danger btn-sm"
                  onClick={() => handleRemove(s.id)}
                  disabled={removingId === s.id}
                >
                  {removingId === s.id ? 'Removing…' : 'Remove'}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      {lyricsFor && (
        <LyricsModal
          lrclibId={lyricsFor.id}
          trackName={lyricsFor.trackName}
          artistName={lyricsFor.artistName}
          onClose={() => setLyricsFor(null)}
        />
      )}

      {addModalOpen && (
        <AddSongModal
          playlistId={id}
          onAdded={load}
          onClose={() => setAddModalOpen(false)}
        />
      )}
    </div>
  );
}

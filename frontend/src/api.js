const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options.headers },
  });
  const text = await res.text();
  if (!res.ok) {
    let errMsg = res.statusText;
    try {
      const data = JSON.parse(text);
      if (data.error) errMsg = data.error;
    } catch (_) {}
    throw new Error(errMsg);
  }
  return text ? JSON.parse(text) : null;
}

export async function searchSongs(q) {
  return request(`/api/songs/search?q=${encodeURIComponent(q)}`);
}

export async function getLyrics(lrclibId) {
  return request(`/api/songs/lyrics/${lrclibId}`);
}

export async function listPlaylists() {
  return request('/api/playlists');
}

export async function getPlaylist(id) {
  return request(`/api/playlists/${id}`);
}

export async function deletePlaylist(id) {
  return request(`/api/playlists/${id}`, { method: 'DELETE' });
}

export async function createPlaylist(name, description = '') {
  return request('/api/playlists', {
    method: 'POST',
    body: JSON.stringify({ name, description }),
  });
}

export async function addSongToPlaylist(playlistId, lrclibId, rating = 0, notes = '') {
  return request(`/api/playlists/${playlistId}/songs`, {
    method: 'POST',
    body: JSON.stringify({ lrclib_id: lrclibId, rating, notes }),
  });
}

export async function removeSongFromPlaylist(playlistId, songId) {
  return request(`/api/playlists/${playlistId}/songs/${songId}`, {
    method: 'DELETE',
  });
}

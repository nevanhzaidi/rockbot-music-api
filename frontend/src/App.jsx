import { Routes, Route, Link, NavLink } from 'react-router-dom';
import Search from './pages/Search';
import Playlists from './pages/Playlists';
import PlaylistDetail from './pages/PlaylistDetail';
import './App.css';

function App() {
  return (
    <div className="app">
      <nav className="nav">
        <Link to="/" className="nav-logo">Rockbot Music</Link>
        <div className="nav-links">
          <NavLink to="/" end>Search</NavLink>
          <NavLink to="/playlists">Playlists</NavLink>
        </div>
      </nav>
      <main className="main">
        <Routes>
          <Route path="/" element={<Search />} />
          <Route path="/playlists" element={<Playlists />} />
          <Route path="/playlists/:id" element={<PlaylistDetail />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;

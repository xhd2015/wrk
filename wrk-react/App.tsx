import { BrowserRouter, Link, Route, Routes } from 'react-router-dom';
import { getRoutePrefix } from 'dot-pkgs-react/routePrefix';
import Home from './pages/Home';
import RepoViewMockPage from './mockup/repo-view/page';
import './styles/app.css';

const routePrefix = getRoutePrefix();

export default function App() {
  return (
    <BrowserRouter basename={routePrefix || undefined}>
      <div className="wrkweb-shell">
        <nav className="wrkweb-nav" aria-label="wrk web">
          <Link to="/" className="wrkweb-brand">
            wrk
          </Link>
          <Link to="/">Home</Link>
          <Link to="/mockup/repo-view">Repo view (mockup)</Link>
        </nav>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/mockup/repo-view" element={<RepoViewMockPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

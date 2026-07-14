import { Link } from 'react-router-dom';

/** Home workflow overview (replaces retired wrkcli/web/page.html). */
export default function Home() {
  return (
    <div className="wrkweb-page">
      <header>
        <h1>
          <span>wrk</span> workflow overview
        </h1>
        <p className="sub">
          Conceptual diagram of how wrk turns a <strong>task</strong> into local{' '}
          <strong>changes</strong>, linked <strong>worktrees</strong>, sync/rebase, and eventually{' '}
          <strong>Main</strong> → tag/push → <strong>Remote</strong>.
        </p>
      </header>
      <main className="home-panel">
        <div className="home-flow" aria-label="wrk workflow diagram">
          <div className="home-node">
            <strong>task</strong>
            <div className="hint">agent-run / prompt</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node">
            <strong>changes</strong>
            <div className="hint">edit in worktree</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node">
            <strong>worktrees</strong>
            <div className="hint">linked checkouts</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node">
            <strong>sync</strong>
            <div className="hint">rebase / merge</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node main">
            <strong>Main</strong>
            <div className="hint">primary branch</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node">
            <strong>tag / push</strong>
            <div className="hint">publish</div>
          </div>
          <div className="home-arrow" aria-hidden="true">
            →
          </div>
          <div className="home-node remote">
            <strong>Remote</strong>
            <div className="hint">upstream</div>
          </div>
        </div>
        <p className="home-legend">
          Local UI for <code>wrk --web</code>. API is mounted at <code>/api/wrk</code> (e.g.{' '}
          <code>GET /api/wrk/projects</code>). See the full repo canvas at{' '}
          <Link to="/mockup/repo-view">/mockup/repo-view</Link>.
        </p>
      </main>
    </div>
  );
}

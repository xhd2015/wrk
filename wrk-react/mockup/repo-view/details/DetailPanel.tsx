import type { ReactNode } from 'react';
import type { LastRunInfo, NodeDetail } from './types';

type Props = {
  detail: NodeDetail | null;
  lastRunByNodeId: Record<string, LastRunInfo>;
  onClose: () => void;
};

export default function DetailPanel({ detail, lastRunByNodeId, onClose }: Props) {
  return (
    <section className="rv-detail" aria-label="object detail" data-testid="detail-panel">
      {!detail ? (
        <div className="rv-detail-empty" data-detail-kind="empty">
          <p className="rv-detail-empty-title">Click a node on the graph to inspect it.</p>
          <p className="rv-detail-empty-hint">
            Entities (Main, worktree, Remote, task) show structure. Actions (tag, push, …) show
            what will run and last-run status. Logs stream below when an action runs.
          </p>
        </div>
      ) : (
        <div
          className={`rv-detail-card rv-detail-card--${detail.role} rv-detail-card--${detail.kind}`}
          data-detail-kind={detail.kind}
          data-detail-node={detail.nodeId}
        >
          <header className="rv-detail-header">
            <div className="rv-detail-heading">
              <span className={`rv-detail-mark rv-detail-mark--${detail.kind}`} aria-hidden="true" />
              <div>
                <div className="rv-detail-title-row">
                  <h2 className="rv-detail-title">{detail.title}</h2>
                  <span className={`rv-detail-badge rv-detail-badge--${detail.role}`}>
                    {detail.role.toUpperCase()}
                    <span className="rv-detail-badge-kind"> · {detail.kind}</span>
                  </span>
                </div>
                <p className="rv-detail-subtitle">{detail.subtitle}</p>
              </div>
            </div>
            <button type="button" className="rv-detail-close" onClick={onClose} aria-label="Close detail">
              ×
            </button>
          </header>
          <div className="rv-detail-body">{renderBody(detail, lastRunByNodeId[detail.nodeId])}</div>
        </div>
      )}
    </section>
  );
}

function renderBody(detail: NodeDetail, lastRun?: LastRunInfo) {
  switch (detail.kind) {
    case 'task':
      return (
        <>
          <Field label="Description">
            <p className="rv-detail-prose">{detail.description}</p>
          </Field>
          <div className="rv-detail-grid">
            <Field label="Slug">
              <code>{detail.slug}</code>
            </Field>
            <Field label="Status">
              <StatusPill status={detail.status} />
            </Field>
            <Field label="Owner">{detail.owner}</Field>
            <Field label="Created">{detail.createdAt}</Field>
          </div>
          {detail.linkedWorktreeLabel ? (
            <Field label="Linked worktree">
              {detail.linkedWorktreeLabel}
              {detail.linkedWorktreeId ? (
                <span className="rv-detail-muted"> ({detail.linkedWorktreeId})</span>
              ) : null}
            </Field>
          ) : null}
          <p className="rv-detail-footnote">Mock fixture only — not a live issue tracker.</p>
        </>
      );

    case 'worktree':
      return (
        <>
          <Field label="Path">
            <code className="rv-detail-path">{detail.path}</code>
          </Field>
          <div className="rv-detail-grid">
            <Field label="Branch">
              <code>{detail.branch}</code>
            </Field>
            <Field label="Base (Main)">
              <code>{detail.baseBranch}</code>
            </Field>
            <Field label="Clean">
              {detail.clean ? (
                <span className="rv-pill rv-pill--ok">clean</span>
              ) : (
                <span className="rv-pill rv-pill--warn">
                  dirty{detail.dirtyFileCount != null ? ` (${detail.dirtyFileCount} files)` : ''}
                </span>
              )}
            </Field>
            <Field label="vs Main">
              ahead {detail.ahead} · behind {detail.behind}
            </Field>
          </div>
          {detail.taskSlug ? (
            <Field label="Task">
              {detail.taskSlug}
              {detail.taskId ? <span className="rv-detail-muted"> ({detail.taskId})</span> : null}
            </Field>
          ) : null}
          <Field label="In-box actions">
            <span className="rv-detail-muted">rebase — contained in worktree frame on graph</span>
          </Field>
        </>
      );

    case 'main':
      return (
        <>
          <Field label="Path">
            <code className="rv-detail-path">{detail.path}</code>
          </Field>
          <div className="rv-detail-grid">
            <Field label="Branch">
              <code>{detail.branch}</code>
            </Field>
            <Field label="Commit">
              <code>{detail.commit}</code>
            </Field>
            <Field label="Status">
              {detail.clean ? (
                <span className="rv-pill rv-pill--ok">clean</span>
              ) : (
                <span className="rv-pill rv-pill--warn">dirty</span>
              )}
            </Field>
            <Field label="Linked worktrees">
              {detail.worktreeCount} ({detail.worktreeLabels.join(', ')})
            </Field>
          </div>
          <Field label="Subject">{detail.subject}</Field>
          <Field label="Downstream on graph">
            <span className="rv-detail-muted">tag → push → Remote</span>
          </Field>
        </>
      );

    case 'remote':
      return (
        <>
          <div className="rv-detail-grid">
            <Field label="Name">
              <code>{detail.name}</code>
            </Field>
            <Field label="Default branch">
              <code>{detail.defaultBranch}</code>
            </Field>
          </div>
          <Field label="URL">
            <code className="rv-detail-path">{detail.url}</code>
          </Field>
          <div className="rv-detail-grid">
            <Field label="Last fetch (mock)">{formatTime(detail.lastFetchAt)}</Field>
            <Field label="Tracking">{detail.trackingSummary}</Field>
          </div>
        </>
      );

    case 'changes':
      return (
        <>
          <div className="rv-detail-grid">
            <Field label="Scope">{detail.scopeLabel}</Field>
            <Field label="Summary">{detail.summary}</Field>
          </div>
          {detail.dirtyFiles.length === 0 ? (
            <p className="rv-detail-prose rv-detail-muted">No dirty files in this mock fixture.</p>
          ) : (
            <table className="rv-detail-table">
              <thead>
                <tr>
                  <th>Path</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {detail.dirtyFiles.map((f) => (
                  <tr key={f.path}>
                    <td>
                      <code>{f.path}</code>
                    </td>
                    <td>
                      <code>{f.status}</code>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      );

    case 'action':
      return (
        <>
          <Field label="Description">
            <p className="rv-detail-prose">{detail.description}</p>
          </Field>
          <div className="rv-detail-grid">
            <Field label="Target">{detail.targetLabel}</Field>
            <Field label="Action">
              <code>{detail.action}</code>
            </Field>
          </div>
          {detail.params && detail.params.length > 0 ? (
            <div className="rv-detail-grid">
              {detail.params.map((p) => (
                <Field key={p.key} label={p.key}>
                  {p.value}
                </Field>
              ))}
            </div>
          ) : null}
          <Field label="What happens (mock)">
            <ol className="rv-detail-steps">
              {detail.steps.map((s) => (
                <li key={s}>{s}</li>
              ))}
            </ol>
          </Field>
          <LastRunBlock lastRun={lastRun} />
          <p className="rv-detail-footnote">
            Hint: click ▸ {detail.title} on the graph to run; logs stream below.
          </p>
        </>
      );

    case 'agent-run':
      return (
        <>
          <div className="rv-detail-grid">
            <Field label="Runner">
              <code>{detail.runner}</code>
            </Field>
            <Field label="Worktree">{detail.worktreeLabel}</Field>
          </div>
          <Field label="Prompt preview (mock)">
            <p className="rv-detail-prose">{detail.promptPreview}</p>
          </Field>
          <Field label="Session">
            <span className="rv-detail-muted">{detail.sessionHint}</span>
          </Field>
          <LastRunBlock lastRun={lastRun} />
        </>
      );

    default:
      return null;
  }
}

function Field({
  label,
  children,
}: {
  label: string;
  children?: ReactNode;
}) {
  return (
    <div className="rv-detail-field">
      <div className="rv-detail-field-label">{label}</div>
      <div className="rv-detail-field-value">{children}</div>
    </div>
  );
}

function StatusPill({ status }: { status: string }) {
  const cls =
    status === 'done' ? 'rv-pill--ok' : status === 'in_progress' ? 'rv-pill--info' : 'rv-pill--muted';
  return <span className={`rv-pill ${cls}`}>{status.replace('_', ' ')}</span>;
}

function LastRunBlock({ lastRun }: { lastRun?: LastRunInfo }) {
  if (!lastRun) {
    return (
      <Field label="Last run">
        <span className="rv-detail-muted">(none yet)</span>
      </Field>
    );
  }
  return (
    <Field label="Last run">
      <code>{lastRun.opId}</code>
      {' · '}
      <span
        className={
          lastRun.status === 'ok'
            ? 'rv-pill rv-pill--ok'
            : lastRun.status === 'running'
              ? 'rv-pill rv-pill--info'
              : 'rv-pill rv-pill--warn'
        }
      >
        {lastRun.status}
      </span>
      {' · '}
      {formatTime(lastRun.at)}
      {lastRun.summary ? (
        <>
          {' · '}
          <span className="rv-detail-muted">{lastRun.summary}</span>
        </>
      ) : null}
      <span className="rv-detail-muted"> (see Logs)</span>
    </Field>
  );
}

function formatTime(ts: string): string {
  try {
    const d = new Date(ts);
    if (Number.isNaN(d.getTime())) return ts;
    return d.toLocaleString(undefined, { hour12: false });
  } catch {
    return ts;
  }
}

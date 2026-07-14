import { useEffect, useRef } from 'react';
import type { OpStatus, UiLogLine } from './model';

type Props = {
  lines: UiLogLine[];
  status: OpStatus;
  opLabel: string | null;
  autoScroll: boolean;
  onClear: () => void;
};

export default function LogPanel({ lines, status, opLabel, autoScroll, onClear }: Props) {
  const scrollerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!autoScroll) return;
    const el = scrollerRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [lines, autoScroll, status]);

  return (
    <section className="rv-log-panel" aria-label="operation logs">
      <header className="rv-log-header">
        <div className="rv-log-title">
          <strong>Logs</strong>
          <span className={`rv-log-status rv-log-status--${status}`}>
            {status}
            {opLabel ? ` · ${opLabel}` : ''}
            {status === 'running' ? ' · streaming…' : ''}
          </span>
        </div>
        <button type="button" className="rv-log-clear" onClick={onClear} disabled={lines.length === 0}>
          Clear
        </button>
      </header>
      <div className="rv-log-body" ref={scrollerRef} data-testid="log-panel-body">
        {lines.length === 0 ? (
          <p className="rv-log-placeholder">Click an action (rebase, sync, tag, push, agent-run) to stream logs…</p>
        ) : (
          <ul className="rv-log-lines">
            {lines.map((line) => (
              <li key={line.id} className={`rv-log-line rv-log-line--${line.level}`}>
                <time dateTime={line.ts}>{formatTime(line.ts)}</time>
                <span className="rv-log-level">{line.level}</span>
                <span className="rv-log-msg">{line.message}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}

function formatTime(ts: string): string {
  try {
    const d = new Date(ts);
    if (Number.isNaN(d.getTime())) return ts;
    return d.toLocaleTimeString(undefined, { hour12: false });
  } catch {
    return ts;
  }
}

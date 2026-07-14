import { useCallback, useMemo, useRef, useState } from 'react';
import { startOp, streamOpLogs, type OpAction } from 'dot-pkgs-react/api/ops';
import DetailPanel from './details/DetailPanel';
import { getDetail } from './details/fixtures';
import type { LastRunInfo } from './details/types';
import LogPanel from './LogPanel';
import RepoViewBoard from './RepoViewBoard';
import { mockRepoView, type OpStatus, type RepoViewNode, type UiLogLine } from './model';
import './repo-view.css';

let logSeq = 0;

/**
 * Interactive mockup at /mockup/repo-view.
 * Graph → kind-specific detail strip → SSE logs.
 */
export default function RepoViewMockPage() {
  const model = mockRepoView;
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [activeNodeId, setActiveNodeId] = useState<string | null>(null);
  const [status, setStatus] = useState<OpStatus>('idle');
  const [opLabel, setOpLabel] = useState<string | null>(null);
  const [lines, setLines] = useState<UiLogLine[]>([]);
  const [lastRunByNodeId, setLastRunByNodeId] = useState<Record<string, LastRunInfo>>({});
  const closeStreamRef = useRef<(() => void) | null>(null);

  const selectedDetail = useMemo(() => getDetail(selectedNodeId), [selectedNodeId]);

  const appendSeparator = useCallback((message: string) => {
    setLines((prev) => [
      ...prev,
      {
        id: `sep-${++logSeq}`,
        ts: new Date().toISOString(),
        level: 'info',
        message: `── ${message} ──`,
      },
    ]);
  }, []);

  const appendLine = useCallback((line: UiLogLine) => {
    setLines((prev) => [...prev, line]);
  }, []);

  const onClear = useCallback(() => {
    setLines([]);
  }, []);

  const onCloseDetail = useCallback(() => {
    setSelectedNodeId(null);
  }, []);

  const onNodeActivate = useCallback(
    async (node: RepoViewNode) => {
      setSelectedNodeId(node.id);

      // Entities / state: detail only (no log spam).
      if (!node.action) {
        return;
      }

      if (status === 'running') {
        appendLine({
          id: `ui-${++logSeq}`,
          ts: new Date().toISOString(),
          level: 'warn',
          message: 'ignored click — an operation is already running',
        });
        return;
      }

      closeStreamRef.current?.();
      closeStreamRef.current = null;

      const action = node.action as OpAction;
      const label = `${action}${node.targetId ? ` @ ${node.targetId}` : ''}`;
      appendSeparator(`start ${label}`);
      setActiveNodeId(node.id);
      setStatus('running');
      setOpLabel(label);
      setLastRunByNodeId((prev) => ({
        ...prev,
        [node.id]: {
          opId: '…',
          status: 'running',
          at: new Date().toISOString(),
        },
      }));

      try {
        const created = await startOp({
          action,
          target_id: node.targetId,
          label: node.label,
        });
        appendLine({
          id: `ui-${++logSeq}`,
          ts: new Date().toISOString(),
          level: 'info',
          message: `op started: ${created.op_id}`,
        });
        setLastRunByNodeId((prev) => ({
          ...prev,
          [node.id]: {
            opId: created.op_id,
            status: 'running',
            at: new Date().toISOString(),
          },
        }));

        const close = streamOpLogs(created.op_id, {
          onLog: (dto) => {
            appendLine({
              id: `log-${++logSeq}`,
              ts: dto.ts,
              level: dto.level,
              message: dto.message,
            });
          },
          onDone: (done) => {
            const summary = done.ok ? done.summary || 'done' : done.error || 'operation failed';
            appendLine({
              id: `done-${++logSeq}`,
              ts: new Date().toISOString(),
              level: done.ok ? 'info' : 'error',
              message: summary,
            });
            setStatus(done.ok ? 'ok' : 'error');
            setActiveNodeId(null);
            closeStreamRef.current = null;
            setLastRunByNodeId((prev) => ({
              ...prev,
              [node.id]: {
                opId: created.op_id,
                status: done.ok ? 'ok' : 'error',
                at: new Date().toISOString(),
                summary,
              },
            }));
          },
          onError: (err) => {
            appendLine({
              id: `err-${++logSeq}`,
              ts: new Date().toISOString(),
              level: 'error',
              message: err.message,
            });
            setStatus('error');
            setActiveNodeId(null);
            closeStreamRef.current = null;
            setLastRunByNodeId((prev) => ({
              ...prev,
              [node.id]: {
                opId: created.op_id,
                status: 'error',
                at: new Date().toISOString(),
                summary: err.message,
              },
            }));
          },
        });
        closeStreamRef.current = close;
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        appendLine({
          id: `err-${++logSeq}`,
          ts: new Date().toISOString(),
          level: 'error',
          message: msg,
        });
        setStatus('error');
        setActiveNodeId(null);
        setLastRunByNodeId((prev) => ({
          ...prev,
          [node.id]: {
            opId: '—',
            status: 'error',
            at: new Date().toISOString(),
            summary: msg,
          },
        }));
      }
    },
    [appendLine, appendSeparator, status],
  );

  return (
    <div className="wrkweb-page repo-view-page">
      <header>
        <h1>
          <span>wrk</span> {model.title ?? 'Repo view'}
        </h1>
        <p className="sub">
          Pure HTML/CSS mock of parallel <strong>task</strong> → <strong>changes</strong> →{' '}
          <strong>worktree</strong> streams (with <strong>agent-run</strong> /{' '}
          <strong>rebase</strong>), <strong>sync</strong> into <strong>Main</strong>, then{' '}
          <strong>tag</strong> → <strong>push</strong> → <strong>Remote</strong>. Click a node for a
          kind-specific detail panel; click actions to stream mock logs.
        </p>
      </header>

      <RepoViewBoard
        model={model}
        selectedNodeId={selectedNodeId}
        activeNodeId={activeNodeId}
        running={status === 'running'}
        onNodeActivate={onNodeActivate}
      />

      <DetailPanel
        detail={selectedDetail}
        lastRunByNodeId={lastRunByNodeId}
        onClose={onCloseDetail}
      />

      <LogPanel
        lines={lines}
        status={status}
        opLabel={opLabel}
        autoScroll
        onClear={onClear}
      />
    </div>
  );
}

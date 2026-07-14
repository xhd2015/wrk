import type { RepoViewModel, RepoViewNode } from './model';

type Props = {
  model: RepoViewModel;
  selectedNodeId: string | null;
  activeNodeId: string | null;
  running: boolean;
  onNodeActivate: (node: RepoViewNode) => void;
};

function nodeRole(kind: RepoViewNode['kind']): 'entity' | 'action' | 'state' {
  switch (kind) {
    case 'worktree':
    case 'main':
    case 'remote':
    case 'task':
      return 'entity';
    case 'action':
    case 'agent-run':
      return 'action';
    default:
      return 'state'; // changes
  }
}

function NodeChip({
  node,
  selected,
  active,
  running,
  onActivate,
}: {
  node: RepoViewNode;
  selected: boolean;
  active: boolean;
  running: boolean;
  onActivate: (n: RepoViewNode) => void;
}) {
  const role = nodeRole(node.kind);
  const className = [
    'rv-node',
    `rv-node--${node.kind}`,
    `rv-node--role-${role}`,
    node.interactive ? 'rv-node--interactive' : '',
    selected ? 'rv-node--selected' : '',
    active ? 'rv-node--active' : '',
  ]
    .filter(Boolean)
    .join(' ');

  const body = (
    <>
      {role === 'entity' ? <span className="rv-node-mark" aria-hidden="true" /> : null}
      {role === 'action' ? <span className="rv-node-verb" aria-hidden="true">▸</span> : null}
      <span className="rv-node-text">{node.label}</span>
    </>
  );

  if (node.interactive) {
    return (
      <button
        type="button"
        className={className}
        data-role={role}
        aria-pressed={active}
        disabled={running && !!node.action}
        onClick={() => onActivate(node)}
      >
        {body}
      </button>
    );
  }
  return (
    <div className={className} data-node-id={node.id} data-role={role}>
      {body}
    </div>
  );
}

/** Pure HTML/CSS conceptual repo board (no SVG). */
export default function RepoViewBoard({
  model,
  selectedNodeId,
  activeNodeId,
  running,
  onNodeActivate,
}: Props) {
  const chip = (node: RepoViewNode) => (
    <NodeChip
      key={node.id}
      node={node}
      selected={selectedNodeId === node.id}
      active={activeNodeId === node.id}
      running={running}
      onActivate={onNodeActivate}
    />
  );

  return (
    <div className="rv-board" data-testid="repo-view-board">
      <div className="rv-board-inner">
        <div className="rv-land rv-land--worktree" aria-label="worktree land">
          {model.streams.map((stream) => (
            <div key={stream.id} className="rv-stream">
              <div className="rv-stream-row">
                <div className="rv-col rv-col--task">
                  {chip(stream.task)}
                  {stream.agentRun ? (
                    <div className="rv-stream-agent">{chip(stream.agentRun)}</div>
                  ) : null}
                </div>
                <span className="rv-arrow" aria-hidden="true">
                  →
                </span>
                {chip(stream.changes)}
                <span className="rv-arrow rv-arrow--up" aria-hidden="true">
                  ↑
                </span>
                <div
                  className={[
                    'rv-worktree-box',
                    selectedNodeId === stream.worktree.id || selectedNodeId === stream.rebase.id
                      ? 'rv-worktree-box--selected'
                      : '',
                    activeNodeId === stream.worktree.id || activeNodeId === stream.rebase.id
                      ? 'rv-worktree-box--active'
                      : '',
                  ]
                    .filter(Boolean)
                    .join(' ')}
                >
                  {stream.worktree.interactive ? (
                    <button
                      type="button"
                      className="rv-worktree-label rv-worktree-label--btn"
                      onClick={() => onNodeActivate(stream.worktree)}
                    >
                      {stream.worktree.label}
                    </button>
                  ) : (
                    <span className="rv-worktree-label">{stream.worktree.label}</span>
                  )}
                  <NodeChip
                    node={stream.rebase}
                    selected={selectedNodeId === stream.rebase.id}
                    active={activeNodeId === stream.rebase.id}
                    running={running}
                    onActivate={onNodeActivate}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="rv-divider" role="separator" aria-orientation="vertical" />

        <div className="rv-land rv-land--mainline" aria-label="mainline publish">
          <div className="rv-publish-row">
            {chip(model.sync)}
            <span className="rv-arrow" aria-hidden="true">
              →
            </span>
            <div className="rv-main-stack">
              {chip(model.main)}
              <div className="rv-main-changes">
                <span className="rv-arrow rv-arrow--up" aria-hidden="true">
                  ↑
                </span>
                {chip(model.mainChanges)}
              </div>
            </div>
            <span className="rv-arrow" aria-hidden="true">
              →
            </span>
            {chip(model.tag)}
            <span className="rv-arrow" aria-hidden="true">
              →
            </span>
            {chip(model.push)}
            <span className="rv-arrow" aria-hidden="true">
              →
            </span>
            {chip(model.remote)}
          </div>
        </div>
      </div>
    </div>
  );
}

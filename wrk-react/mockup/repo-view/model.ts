import type { OpAction } from 'dot-pkgs-react/api/ops';

export type NodeKind =
  | 'task'
  | 'agent-run'
  | 'changes'
  | 'worktree'
  | 'action'
  | 'main'
  | 'remote';

export type RepoViewNode = {
  id: string;
  kind: NodeKind;
  label: string;
  /** When true, render as button and can start a mock op / selection. */
  interactive?: boolean;
  /** Backend mock action for interactive action nodes. */
  action?: OpAction;
  /** Optional target for the op (e.g. worktree id). */
  targetId?: string;
};

export type WorktreeStream = {
  id: string;
  task: RepoViewNode;
  agentRun?: RepoViewNode;
  changes: RepoViewNode;
  worktree: RepoViewNode;
  rebase: RepoViewNode;
};

export type RepoViewModel = {
  id: string;
  title?: string;
  streams: WorktreeStream[];
  sync: RepoViewNode;
  main: RepoViewNode;
  mainChanges: RepoViewNode;
  tag: RepoViewNode;
  push: RepoViewNode;
  remote: RepoViewNode;
};

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export type UiLogLine = {
  id: string;
  ts: string;
  level: LogLevel | string;
  message: string;
};

export type OpStatus = 'idle' | 'running' | 'ok' | 'error';

/** Hard-coded fixture matching the conceptual repo-view sketch. */
export const mockRepoView: RepoViewModel = {
  id: 'mock-repo-view',
  title: 'Repo view (mockup)',
  streams: [
    {
      id: 'stream-1',
      task: { id: 'task-1', kind: 'task', label: 'task', interactive: true },
      agentRun: {
        id: 'agent-1',
        kind: 'agent-run',
        label: 'agent-run',
        interactive: true,
        action: 'agent-run',
        targetId: 'worktree-1',
      },
      changes: { id: 'chg-1', kind: 'changes', label: 'changes', interactive: true },
      worktree: {
        id: 'wt-1',
        kind: 'worktree',
        label: 'worktree 1',
        interactive: true,
      },
      rebase: {
        id: 'rebase-1',
        kind: 'action',
        label: 'rebase',
        interactive: true,
        action: 'rebase',
        targetId: 'worktree-1',
      },
    },
    {
      id: 'stream-2',
      task: { id: 'task-2', kind: 'task', label: 'task', interactive: true },
      agentRun: {
        id: 'agent-2',
        kind: 'agent-run',
        label: 'agent-run',
        interactive: true,
        action: 'agent-run',
        targetId: 'worktree-2',
      },
      changes: { id: 'chg-2', kind: 'changes', label: 'changes', interactive: true },
      worktree: {
        id: 'wt-2',
        kind: 'worktree',
        label: 'worktree 2',
        interactive: true,
      },
      rebase: {
        id: 'rebase-2',
        kind: 'action',
        label: 'rebase',
        interactive: true,
        action: 'rebase',
        targetId: 'worktree-2',
      },
    },
  ],
  sync: {
    id: 'sync',
    kind: 'action',
    label: 'sync',
    interactive: true,
    action: 'sync',
    targetId: 'main',
  },
  main: { id: 'main', kind: 'main', label: 'Main', interactive: true },
  mainChanges: { id: 'chg-main', kind: 'changes', label: 'changes', interactive: true },
  tag: {
    id: 'tag',
    kind: 'action',
    label: 'tag',
    interactive: true,
    action: 'tag',
    targetId: 'main',
  },
  push: {
    id: 'push',
    kind: 'action',
    label: 'push',
    interactive: true,
    action: 'push',
    targetId: 'main',
  },
  remote: { id: 'remote', kind: 'remote', label: 'Remote', interactive: true },
};

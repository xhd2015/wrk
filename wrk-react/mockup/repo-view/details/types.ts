import type { OpAction } from 'dot-pkgs-react/api/ops';

export type DetailRole = 'entity' | 'state' | 'action';

export type DetailBase = {
  nodeId: string;
  title: string;
  subtitle: string;
  role: DetailRole;
};

export type TaskDetail = DetailBase & {
  kind: 'task';
  description: string;
  slug: string;
  status: 'open' | 'in_progress' | 'done';
  owner: string;
  createdAt: string;
  linkedWorktreeId?: string;
  linkedWorktreeLabel?: string;
};

export type WorktreeDetail = DetailBase & {
  kind: 'worktree';
  path: string;
  branch: string;
  baseBranch: string;
  clean: boolean;
  dirtyFileCount?: number;
  ahead: number;
  behind: number;
  taskId?: string;
  taskSlug?: string;
};

export type MainDetail = DetailBase & {
  kind: 'main';
  path: string;
  branch: string;
  commit: string;
  subject: string;
  clean: boolean;
  worktreeCount: number;
  worktreeLabels: string[];
};

export type RemoteDetail = DetailBase & {
  kind: 'remote';
  name: string;
  url: string;
  defaultBranch: string;
  lastFetchAt: string;
  trackingSummary: string;
};

export type ChangesDetail = DetailBase & {
  kind: 'changes';
  scope: 'worktree' | 'main';
  scopeLabel: string;
  dirtyFiles: { path: string; status: string }[];
  summary: string;
};

export type ActionDetail = DetailBase & {
  kind: 'action';
  action: OpAction;
  targetLabel: string;
  description: string;
  steps: string[];
  params?: { key: string; value: string }[];
};

export type AgentRunDetail = DetailBase & {
  kind: 'agent-run';
  action: 'agent-run';
  runner: string;
  worktreeLabel: string;
  promptPreview: string;
  sessionHint: string;
};

export type NodeDetail =
  | TaskDetail
  | WorktreeDetail
  | MainDetail
  | RemoteDetail
  | ChangesDetail
  | ActionDetail
  | AgentRunDetail;

export type LastRunInfo = {
  opId: string;
  status: 'ok' | 'error' | 'running';
  at: string;
  summary?: string;
};

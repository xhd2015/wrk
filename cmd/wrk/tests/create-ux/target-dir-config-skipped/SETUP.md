# Scenario

**Feature**: create-with-target-dir skips preconfigured create UX from config.json

```
# spawnTarget non-empty → config create.* is NOT base-merged
myrepo + <target-dir> + full create.* config
  -> native create at/under target; no space / iterm / agent from config

# CLI one-shot UX flags still apply with <target-dir>
wrk <dir> <target-dir> --new-window|--new-terminal|--open-in-agent
  -> UX from flags only (empty config base)

# window-implies-terminal-new still after flag apply
```

## Preconditions

- Git available; isolated `{WRK_HOME}`; `WRK_DATE=2026-06-30`.
- Same hermetic UX mocks as sibling create-ux groups (`installCreateUXMocks`).
- Source repo is `myrepo` under `{WorkRoot}`; process cwd is `{WorkRoot}` so absolute
  `<target-dir>` is independent of repo path.
- Default spawn shape: target missing, parent exists → worktree exactly at `SpawnDir`
  (no naming suffix on path; not under `{WRK_HOME}/worktrees`).

## Steps

- Group Setup: seed main repo, install darwin UX mocks, set
  `TargetDir` + `SpawnDir` for `wrk <myrepo> <wt>`.
- Leaves write config and/or CLI UX flags; assert path + mock side effects.

## Context

- `req.TargetDir` = absolute source repo (first positional).
- `req.SpawnDir` = absolute `<target-dir>` under `{WorkRoot}/wt` (second positional).
- `req.RepoDir` = `{WorkRoot}` (process cwd for the wrk invocation).
- Contrast with `pipeline/config/defaults-match-flags`: same full config without
  `SpawnDir` **does** drive space/iterm/agent.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	mainRepo := setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	// wrk <dir> <target-dir>: first positional = source, second = spawn override.
	req.TargetDir = mainRepo
	req.RepoDir = req.WorkRoot
	// Missing target, parent WorkRoot exists → exact path spawn.
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	return nil
}
```

# Scenario

**Feature**: wrk from main repo checkout

```
# cwd is the primary git checkout (not a linked worktree)
main repo checkout -> wrk -> {WRK_HOME}/worktrees/{basename}-{token}-{YYYY-MM-DD}
# branch = {token}-{date}[-N]; always worktree add -b (never reuse existing branch)
# token = sanitize(currentBranch) with / → -
```

## Behavior-change leaves (always-new-branch + slash sanitize)

- `slash-branch/` — branch name is tokenized (`feature-foo-{date}`), not slash-preserving.
- `branch-collision/` — pre-existing branch forces joint path+branch `-1` via new `-b` (no reuse).
- `sequence-increment/` — second create always new branch `-1` (C1 regression strengthened).

## Steps

- Tests create a fresh git repo under `{WorkRoot}/myrepo` unless a leaf overrides setup.
- `req.RepoDir` points at the main checkout directory.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	req.RepoDir = repoDir
	return nil
}
```
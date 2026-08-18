# Scenario

**Feature**: named create when the dump path already exists as a linked worktree → error, no create

```
# dump sibling already occupies {WRK_HOME}/worktrees/myrepo-main-{date}
# wrk myrepo <that-path> -> non-zero; already exists
# skip Policy B ≠ reuse this named worktree; must not nest a named child
myrepo (main) + existing linked WT under dump
  -> wrk myrepo {WRK_HOME}/worktrees/myrepo-main-{date}
  -> error; no nested create
```

## Steps

1. Grouping already created a live linked worktree at `{req.WrkHome}/worktrees/myrepo-main-{date}`.
2. Set `req.SpawnDir` to that occupied dump path.
3. Run `wrk myrepo <exact-dump-path>` in-process.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.UseScriptTTY = false
	req.SpawnDir = req.WtDir
	return nil
}
```

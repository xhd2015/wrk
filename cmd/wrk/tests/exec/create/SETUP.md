# Scenario

**Feature**: `--exec` after native create runs the command in the new worktree

```
# create mode + --exec
myrepo (main) -> wrk --exec <cmd>
  -> native worktree under WRK_HOME
  -> stdout: <wt-abs>\n + child stdout
  -> child cmd.Dir = wt-abs
```

## Preconditions

- Empty create UX config (no window/terminal/agent) unless a leaf writes one.
- Success leaves run bare create + `--exec` (native pipeline only).

## Steps

- Default: init `myrepo` on main; `req.RepoDir` = main checkout.
- Leaves set `req.Args` including `--exec …` (or error forms).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	req.RepoDir = repoDir
	req.MainRepo = repoDir
	return nil
}
```

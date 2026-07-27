# Scenario

**Feature**: wrk --done rejects dirty worktree

```
# linked wt with uncommitted changes
myrepo + wt -> touch dirty file -> wrk --done -> error
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Write an uncommitted file in the worktree.
3. Run `wrk --done`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	writeFile(t, filepath.Join(wtDir, "dirty-file"), "uncommitted")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
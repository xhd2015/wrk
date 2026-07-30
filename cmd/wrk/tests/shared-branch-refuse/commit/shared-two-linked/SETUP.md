# Scenario

**Feature**: `wrk --gen-commit-msg --commit` refuses on a shared branch (dry-run hard error)

```
# two live checkouts of B; staged file on wt1
wt1 (B, staged) + wt2 (B)
  -> wrk --gen-commit-msg --commit --dry-run
  -> non-zero; Error: refuse … commit
  -> HEAD subject unchanged; change.go stays staged (no commit)
```

## Steps

1. Shared two-linked + stage `change.go` on primary (no commit of that file).
2. Snapshot HEAD subject into `req.HashToken` (string field reuse for this leaf).
3. Run `wrk --gen-commit-msg --commit --dry-run` from primary wt.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedTwoLinkedStaged(t, req)
	req.HashToken = strings.TrimSpace(gitOutputIsolated(t, req.RepoDir, "log", "-1", "--format=%s"))
	req.Args = []string{"--gen-commit-msg", "--commit", "--dry-run"}
	return nil
}
```

# Scenario

**Feature**: without --main the linked worktree checkout is the scan root (contrast)

```
# I9: cwd=linked-wt, no --main
linked-wt -> wrk --install wtbin --dry-run
  -> useMain=false → scan ShowToplevel(linked-wt)
  -> would: go install ./cmd/wtbin only (K=1)
```

## Steps

1. Parent built diverged main + linked-wt; cwd = linked-wt.
2. Args = `--install wtbin --dry-run` (drop `--main`).
3. Expect the single-module install dry-run for the worktree module.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--install", "wtbin", "--dry-run"}
	return nil
}
```

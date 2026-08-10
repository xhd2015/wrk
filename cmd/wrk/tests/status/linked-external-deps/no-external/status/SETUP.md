# Scenario

**Feature**: `wrk --status` from linked consumer with no external deps is a single block

```
consumer main + wrk --new -> consumerWt
  -> cwd=consumerWt, wrk --status
  -> one Dir: . block with Master:; no ---- external ----
```

## Steps

1. Create consumer main + linked worktree via `wrk --new`.
2. Run `wrk --status` from the linked consumer root (InProcess Capture).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_, consumerWt, _ := setupLinkedConsumer(t, req)
	req.RepoDir = consumerWt
	req.Args = []string{"--status"}
	return nil
}
```

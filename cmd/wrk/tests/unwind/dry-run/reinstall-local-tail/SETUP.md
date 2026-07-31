# Scenario

**Feature**: `--reinstall-local` is an unwind tail request, including in dry-run mode

```
# dirty primary repo; no stack edges and no linked-worktree land required
root (dirty, main) -> wrk --unwind --dry-run --reinstall-local
  -> ordinary unwind peel plan
  -> exit 0; no mutation and no mutual-exclusion rejection
```

## Steps

1. Create a sole dirty main repository with a Go module.
2. Run unwind dry-run with `--reinstall-local`; there are no edges, pins, or land flags.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--reinstall-local"}
	recordUnwindBaseline(t, req)
	return nil
}
```

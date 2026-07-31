# Scenario

**Feature**: single dirty main repo with no stack edges needs no pin flags

```
# main-only root; dirty; already on main → no land; no require/replace edges
root (dirty, main) -> wrk --unwind --dry-run
  -> would: peel root
  -> exit 0 without --tag-next/--push
```

## Steps

1. Create sole main repo `root` with a go.mod and dirty untracked file.
2. Run `--unwind --dry-run` only (no ship/land flags).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = []string{"--unwind", "--dry-run"}
	recordUnwindBaseline(t, req)
	return nil
}
```

# Scenario

**Feature**: cross-repo edges among pending require --tag-next and --push

```
# 3-repo dirty chain; pin flags omitted
all dirty + edges -> wrk --unwind --dry-run
  -> exit ≠ 0
  -> Error names --tag-next and --push
  -> zero mutations (no peel apply)
```

## Steps

1. Build three-repo dirty chain (edges present).
2. Run `--unwind --dry-run` **without** `--tag-next` / `--push` (and without land).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	// Intentionally omit pin flags (and land) — validation must hard-fail.
	req.Args = []string{"--unwind", "--dry-run"}
	recordUnwindBaseline(t, req)
	return nil
}
```

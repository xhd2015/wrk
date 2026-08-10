# Scenario

**Feature**: peel-only dry-run when `--tag-next` is omitted (C-DR3)

```
# same multi-module fixture as C-DR1 but args omit --tag-next
root + shared (dirty)
  -> wrk --unwind --dry-run
  -> would: peel .
  -> no would: tag-next; no cascade pin … <- …
  -> exit 0; zero mutations
```

## Steps

1. Seed single-repo two modules (shared owned-changed).
2. Run `--unwind --dry-run` only.
3. Assert peel plan without cascade module lines.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeSingleRepoTwoModules(t, req)
	req.Args = []string{"--unwind", "--dry-run"}
	recordUnwindBaseline(t, req)
	return nil
}
```

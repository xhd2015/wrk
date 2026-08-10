# Scenario

**Feature**: apply cascade when go.mod/go.sum match Base (clean pin path)

```
# clean Base for go.mod/go.sum (DIRTY peel marker OK); no --add-all required
stack clean mods + --unwind --tag-next [+ --push/--done]
  -> land prelude then free-module cascade
  -> pin selective commit + tags; exit 0
```

## Preconditions

- Parent apply-cascade helpers: `setupApplyCascade*`, pin/tag asserts.
- Leaves do **not** leave uncommitted go.mod/go.sum WIP before Run.

## Steps

1. Grouping locks clean worktree policy for go.mod/go.sum.
2. Leaves split on stack shape (single-repo multi-mod vs multi-repo).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```

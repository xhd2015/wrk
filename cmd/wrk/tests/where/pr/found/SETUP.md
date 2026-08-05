# Scenario

**Feature**: valid `--where --pr` URL resolves to one or more local worktree paths

```
full GitHub PR URL + gh view headRefName + matching main(s)
  -> live worktree(s) on head -> stdout abs path(s) lex-sorted
```

## Preconditions

- Fake `gh` returns headRefName; git fixtures provide matching origin and worktrees.

## Steps

- Leaves vary match count, discovery scope, PR state, and flag order.

## Context

- Success: exit 0, empty stderr, paths only on stdout.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```

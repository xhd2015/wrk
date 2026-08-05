# Scenario

**Feature**: multiple live checkouts of PR head → all paths printed, lex-sorted

```
two mains (clones) of owner/repo, each with head wt
  -> stdout both abs paths sorted, one per line
```

## Preconditions

- Two separate main clones (git forbids two worktrees of one main on the same branch).

## Steps

- Leaves seed multi-clone fixtures.

## Context

- Exit 0; empty stderr.

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

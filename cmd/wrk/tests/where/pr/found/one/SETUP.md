# Scenario

**Feature**: exactly one local worktree on PR head → one path line

```
one live checkout of headRefName -> stdout that abs path + \n
```

## Preconditions

- Single linked worktree on head branch.

## Steps

- Leaves vary how the main is discovered (projects.json vs cwd).

## Context

- Exit 0; stderr empty.

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

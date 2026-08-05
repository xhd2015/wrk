# Scenario

**Feature**: closed/merged PR still resolves location (not open-only filter)

```
gh pr view state=CLOSED, headRefName set
  -> still print local worktree path for head
```

## Preconditions

- Fake view returns non-OPEN state with headRefName.

## Steps

- Leaves set CLOSED (or MERGED) view JSON.

## Context

- Location lookup is not restricted to open PRs.

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

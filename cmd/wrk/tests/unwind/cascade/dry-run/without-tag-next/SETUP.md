# Scenario

**Feature**: without `--tag-next`, dry-run stays peel-only (no cascade) (C-DR3)

```
# multi-module stack but no --tag-next
  -> would: peel …
  -> MUST NOT would: tag-next / cascade would: pin … <- …
```

## Preconditions

- Single-repo (no cross-repo edges) so dry-run succeeds without pin flags.
- Status-quo guard: may be **GREEN** today; must stay green after cascade lands.

## Steps

1. Grouping locks absence of `--tag-next`.

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

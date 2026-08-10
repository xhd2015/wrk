# Scenario

**Feature**: cascade plan under `--unwind --dry-run` (P1 — no apply)

```
# acyclic + --tag-next → peel then free-first would: tag-next / would: pin
# without --tag-next → peel-only (no cascade)
# cycle → reject; no cascade body
wrk --unwind --dry-run [+ --tag-next …] -> plan or cycle error; zero mutations
```

## Preconditions

- Parent cascade helpers: `setupCascade*`, cascade line asserts.
- All leaves include `--dry-run` (no apply side-effect asserts for tag create).

## Steps

1. Grouping marks the dry-run cascade family.
2. Descendants split on `--tag-next` / cycle.

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

# Scenario

**Feature**: stack-repo DAG cycle aborts before any mutation

```
# mutual require edges among nested stack repos
A ↔ B under checkout -> wrk --unwind [--dry-run] [flags]
  -> exit ≠ 0; message mentions cycle
  -> no successful peel plan; zero mutations
```

## Preconditions

- Parent helpers: `setupTwoCycleStack`, `assertCycleError`, baseline asserts.
- Cycle check runs on dry-run **and** apply (preflight before any mutation).

## Steps

1. Grouping scopes cycle-preflight leaves; descendants build cyclic fixtures
   (dry-run leaf and optional apply-mode leaf).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	unwindEnsureHelpersUsed()
	return nil
}
```

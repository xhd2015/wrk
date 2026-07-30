# Scenario

**Feature**: legacy `--dep` and `--all-deps` are hard-removed from the CLI surface

```
# end-state: only --bring remains as external-dep worktree mode
wrk --dep <path> -> non-zero; unknown/invalid flag --dep
wrk --all-deps -> non-zero; unknown/invalid flag --all-deps
# help / dry-run host lists no longer advertise the removed flags
```

## Steps

- Leaves use minimal cwd fixtures and L2 `InProcess` for fast-fail unknown-flag paths.
- Asserts prefer stable substrings: flag name + `unknown` / invalid-style messaging.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	return nil
}
```

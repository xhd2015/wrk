# Scenario

**Feature**: --main --install resolves names from the main repo modules (I8)

```
# I8: cwd=linked-wt (diverged); Args = --main --install mainbin toolbin --dry-run
linked-wt -> wrk --main --install mainbin toolbin --dry-run
  -> useMain=true → scan mainrepo
  -> would: go install ./cmd/mainbin + ./cmd/toolbin (K=2; not wtbin)
  -> no nested shell; dry-run only
```

## Steps

1. Parent built diverged main + linked-wt; cwd = linked-wt.
2. Args default to `--main --install mainbin toolbin --dry-run` (parent Setup).
3. Expect the multi-module install dry-run for **main** modules only.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Parent from-linked-wt sets ModuleRoot=linked-wt and the compose Args.
	req.Args = []string{"--main", "--install", "mainbin", "toolbin", "--dry-run"}
	return nil
}
```

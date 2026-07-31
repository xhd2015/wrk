# Scenario

**Feature**: dry-run exposes every selected per-peel stage

```
dirty linked dependency -> unwind dry-run -> expanded ordered plan, no mutation
```

```go
import "github.com/xhd2015/doctest/session"
func Setup(t *testing.T, d *session.Doctest, req *Request) error { _ = d; seedLinkedDep(t, req); return nil }
```

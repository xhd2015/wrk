# Scenario

**Feature**: dir-mode soft-skips broken nested checkouts discovered under the stack

```
# git consumer + nested checkout with stale gitdir
  -> wrk --dep-update <dep> --dry-run
  -> warning: skipping nested checkout <rel-path>: …
  -> continues; would: pin on primary; exit 0
```

## Steps

1. Leaves seed a stack primary with a broken nested `.git` pointer.
2. Run dir-mode dry-run.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```

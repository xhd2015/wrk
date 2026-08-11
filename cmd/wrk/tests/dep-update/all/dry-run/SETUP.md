# Scenario

**Feature**: wrk --dep-update --all --dry-run plans inventory pins without writes

```
consumer git toplevel + inventory owners
  -> wrk --dep-update --all --dry-run
  -> would: dep-update / would: go mod tidy (when pins planned)
  -> zero go.mod writes
```

## Steps

1. Grouping defaults dry-run Args; leaves seed fixtures.

## Context

- Default args: `[]string{"--dep-update", "--all", "--dry-run"}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	if len(req.Args) == 0 {
		req.Args = []string{"--dep-update", "--all", "--dry-run"}
	}
	ensureDepUpdateHelpersUsed()
	return nil
}
```

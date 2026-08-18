# Scenario

**Feature**: wrk --dep-update --all --dry-run plans inventory pins without writes

```
# stack consumers + inventory owners
  -> wrk --dep-update --all --dry-run
  -> ==== dep-update (dry-run) ====; would: pin / would: go mod tidy (when pins planned)
  -> no argv dep list; zero go.mod writes
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

# Scenario

**Feature**: `-f` / `--force` without `--push` is a hard error

```
# flag layer: force is only valid with --push
wrk -f | wrk --force  (no --push)
  -> non-zero
  -> stderr: wrk: -f/--force is only valid with --push
  -> no push / no confirm line
```

## Steps

- Grouping: leaves set `req.Args` to force flag alone (or with non-push modes if needed later).
- Flag validation only — no origin fixture required.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Flag reject before mode work; isolated WorkRoot is enough as cwd.
	req.RepoDir = req.WorkRoot
	return nil
}
```

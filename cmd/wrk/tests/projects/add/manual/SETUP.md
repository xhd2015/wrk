# Scenario

**Feature**: wrk --add resolves directory to main repo

```
wrk --add <dir> -> projects.json + stdout resolved main repo path
```

## Steps

- Descendants vary whether `<dir>` is main repo or linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureProjectsHelpersUsed()
	return nil
}
```
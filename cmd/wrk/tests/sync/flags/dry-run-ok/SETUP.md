# Scenario

**Feature**: wrk --sync --dry-run succeeds on a main-only git repo

```
# main checkout, no linked worktrees
myrepo (main only) -> wrk --sync --dry-run -> exit 0, would: zero summary
```

## Steps

1. `initMainOnlyRepo`.
2. Run `wrk --sync --dry-run` from the main repo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "--dry-run"}
	return nil
}
```

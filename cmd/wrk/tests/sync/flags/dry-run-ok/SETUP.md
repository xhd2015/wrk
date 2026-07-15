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
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "--dry-run"}
	return nil
}
```

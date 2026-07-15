# Scenario

**Feature**: wrk --sync on main-only repo is a successful no-op with zero counts

```
# single main checkout, no linked named-branch worktrees
myrepo (main only) -> wrk --sync -> exit 0, synced: 0 into main, 0 into worktrees, 0 skipped
```

## Steps

1. `initMainOnlyRepo`.
2. Run `wrk --sync` from the main repo.

```go
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync"}
	return nil
}
```

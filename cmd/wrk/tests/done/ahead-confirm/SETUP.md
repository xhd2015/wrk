# Scenario

**Feature**: wrk --done ff-merges ahead branch by default (no confirm flag / stdin)

```
# wt branch ahead of main; default auto-yes — no --confirm-from-stdin, no Proceed?
myrepo + wt -> commit on wt -> wrk --done -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done` with no confirm flags and no stdin.

```go
func Setup(t *testing.T, req *Request) error {
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```

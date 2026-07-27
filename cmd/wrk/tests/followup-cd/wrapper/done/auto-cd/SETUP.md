# Scenario

**Feature**: wrapper --done from linked worktree cds to main repo

```
wt already-included; source bash.sh; wrk --done
  -> wt removed; stderr cd <main>; FinalPWD = main
```

## Steps

1. Create linked wt; merge into main (already-included).
2. Run `wrk --done` via wrapper from worktree cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.RepoDir = wtDir
	req.StartDir = wtDir
	req.CLIArgs = []string{"--done"}
	return nil
}
```

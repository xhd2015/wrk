# Scenario

**Feature**: --done on a linked worktree while shell cwd is main writes no follow-up

```
# shell cwd = main checkout (still valid after); operate on linked wt
main (cwd) + wt already-included + WRK_FOLLOWUP_FILE
wrk --done <wt> -> wt removed; follow-up empty (main still exists)
```

## Steps

1. Create main repo and one linked worktree; commit and ff-merge into main.
2. Run `wrk --done <wt>` with process cwd = main repo and follow-up env set.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.WtDir = wtDir
	// Process/shell cwd is main checkout, not the operated worktree.
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--done", wtDir}
	return nil
}
```

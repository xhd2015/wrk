# Scenario

**Feature**: successful --done remove writes cd to main repo

```
wrk creates wt; merge branch into main; wrk --done + WRK_FOLLOWUP_FILE
  -> wt removed; follow-up: cd <main-repo>
```

## Steps

1. Create main + linked worktree; commit and ff-merge into main (already-included path).
2. Run `wrk --done` from worktree with follow-up env set.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.RepoDir = wtDir
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--done"}
	return nil
}
```

# Scenario

**Feature**: successful --merge-back with env set writes no follow-up

```
wt already-included; wrk --merge-back -y + WRK_FOLLOWUP_FILE
  -> exit 0; wt kept; follow-up empty
```

## Steps

1. Create linked worktree; commit; ff-merge into main.
2. Run `wrk --merge-back -y` with follow-up env (wt remains).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	req.RepoDir = wtDir
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--merge-back", "-y"}
	return nil
}
```

# Scenario

**Feature**: successful `--done` + `--exec pwd` runs in main repo after remove

```
myrepo + wt (already merged into main)
  -> wrk --done -y --exec pwd
  -> remove wt; stdout includes "worktree removed:" then main path from pwd
  -> wt dir gone; pwd is main, not removed wt
```

## Steps

1. Create main + linked worktree; commit on wt and ff-merge into main (already-included).
2. Run `wrk --done -y --exec pwd` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)

	// Canonical main path for pwd compare (EvalSymlinks matches git/wrk).
	req.MainRepo = resolveAbs(t, mainRepo)
	req.RepoDir = wtDir
	req.Args = []string{"--done", "-y", "--exec", "pwd"}
	return nil
}
```

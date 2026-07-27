# Scenario

**Feature**: wrk --done -v logs merge and worktree remove

```
branch already merged -> wrk --done -v from linked wt
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "already merged")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branch)
	req.RepoDir = wtDir
	req.Args = []string{"--done", "-v"}
	return nil
}
```
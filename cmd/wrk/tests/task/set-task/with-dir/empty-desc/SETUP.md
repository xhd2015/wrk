# Scenario

**Feature**: wrk <dir> --set-task with empty description errors before any rename

```
wrk <linked-worktree-dir> --set-task "" -> non-zero exit, error about empty description
```

## Steps

1. Create a worktree with --task.
2. Run `wrk <wt-dir> --set-task ""`.
3. Verify non-zero exit with error about empty description.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = req.WorkRoot
	req.TargetDir = wtDir
	req.Args = []string{"--set-task", ""}
	return nil
}
```
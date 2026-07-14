# Scenario

**Feature**: --set-task on a non-wrk worktree (custom branch) errors

```
git worktree add -b my-feature ...
wrk --set-task "x" -> cannot parse branch name -> error
```

## Steps

1. Create main repo.
2. Create a linked worktree with custom branch.
3. Set req.SetTaskDesc and run from inside that worktree.
4. Verify non-zero exit.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	// Create a worktree with custom branch name
	wtDir := filepath.Join(req.WorkRoot, "custom-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "my-feature", wtDir)
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "my task"
	return nil
}
```
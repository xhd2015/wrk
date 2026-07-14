# Scenario

**Feature**: --task with target-dir: branch includes slug, dir is user-specified

```
wrk <repo> <target-dir> --task "desc" -> dir is exactly <target-dir>, branch includes slug
```

## Steps

1. Create repo.
2. Use TargetDir + SpawnDir + TaskDesc so Run passes wrk <dir> <target> --task.
3. Verify dir is exactly <target-dir>.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	taskDesc := "fix it"
	slug := slugify(taskDesc)
	req.TaskDesc = taskDesc
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	req.SpawnDir = filepath.Join(req.WorkRoot, "my-custom-dir")
	req.WtBranch = branchNameWithTask("main", wrkDate, slug, 0)
	return nil
}
```
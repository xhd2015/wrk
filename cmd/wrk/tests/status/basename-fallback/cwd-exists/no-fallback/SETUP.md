# Scenario

**Feature**: cwd ./myrepo blocks fallback even when saved project exists for --status

```
WorkRoot/myrepo (non-git dir exists)
saved/myrepo recorded in projects.json
WorkRoot -> wrk myrepo --status -> is not a git repository (no fallback to saved)
```

## Steps

1. Create non-git directory `{WorkRoot}/myrepo`.
2. Create and record saved git repo at `{WorkRoot}/saved/myrepo`.
3. Run `wrk myrepo --status` from `{WorkRoot}`.

```go
func Setup(t *testing.T, req *Request) error {
	initNonGitBasenameInDir(t, req.WorkRoot, "myrepo")
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = req.WorkRoot
	req.TargetDir = "myrepo"
	return nil
}
```

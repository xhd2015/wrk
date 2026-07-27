# Scenario

**Feature**: cwd ./myrepo blocks fallback even when saved project exists

```
WorkRoot/myrepo (non-git dir exists)
saved/myrepo recorded in projects.json
WorkRoot -> wrk myrepo -> not a git repository (no fallback to saved)
```

## Steps

1. Create non-git directory `{WorkRoot}/myrepo`.
2. Create and record saved git repo at `{WorkRoot}/saved/myrepo`.
3. Run `wrk myrepo` from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initNonGitBasenameDir(t, req.WorkRoot, "myrepo")
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = req.WorkRoot
	req.TargetDir = "myrepo"
	return nil
}
```
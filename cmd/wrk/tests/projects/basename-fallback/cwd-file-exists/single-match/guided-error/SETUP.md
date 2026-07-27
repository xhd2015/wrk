# Scenario

**Feature**: wrk myrepo -t emits guided error when cwd file blocks basename resolution

```
workspace/myrepo (file)
saved/myrepo in projects.json
wrk myrepo -t 'optimize skills output' -> multi-line stderr + concrete path hint
```

## Steps

1. Create git repo at `{WorkRoot}/saved/myrepo` and record it.
2. Create neutral cwd `{WorkRoot}/workspace` with regular file `myrepo` (not a directory).
3. Run `wrk myrepo -t 'optimize skills output'` from workspace.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	cwd := initNeutralCwd(t, req.WorkRoot, "workspace")
	initBasenameFile(t, cwd, "myrepo", "")
	req.RepoDir = cwd
	req.TargetDir = "myrepo"
	req.TaskDesc = "optimize skills output"
	req.TaskFlag = "-t"
	return nil
}
```
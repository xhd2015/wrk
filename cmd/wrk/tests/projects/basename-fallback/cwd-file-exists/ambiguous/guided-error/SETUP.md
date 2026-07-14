# Scenario

**Feature**: wrk spl --status emits ambiguous guided error when cwd file blocks resolution

```
workspace/spl (file)
aaa/spl + zzz/spl in projects.json
wrk spl --status -> multi-line stderr + <full-path> hint
```

## Steps

1. Create and record `{WorkRoot}/aaa/spl` and `{WorkRoot}/zzz/spl`.
2. Create neutral cwd `{WorkRoot}/workspace` with regular file `spl`.
3. Run `wrk spl --status` from workspace.

```go
func Setup(t *testing.T, req *Request) error {
	repoA := initSavedGitRepo(t, req.WorkRoot, "aaa", "spl")
	repoZ := initSavedGitRepo(t, req.WorkRoot, "zzz", "spl")
	recordSavedProject(t, req, repoA)
	recordSavedProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	cwd := initNeutralCwd(t, req.WorkRoot, "workspace")
	initBasenameFile(t, cwd, "spl", "")
	req.RepoDir = cwd
	req.TargetDir = "spl"
	req.Args = []string{"--status"}
	return nil
}
```
# Scenario

**Feature**: in-place wrk --cd myrepo expands basename via projects.json

```
saved/myrepo recorded; no ./myrepo under cwd
WRK_FOLLOWUP_FILE set
wrk --cd myrepo -> follow-up cd <saved abs> (never literal "myrepo")
```

## Steps

1. Create and record git repo `{WorkRoot}/saved/myrepo`.
2. Neutral cwd without local `myrepo` dir.
3. Run `wrk --cd myrepo` with channel open.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	saved := initSavedGitRepo(t, req.WorkRoot, "saved", cdBasename)
	recordSavedProject(t, req, saved)
	req.MainRepo = resolvePath(t, saved)
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	setCDFlagThenPath(req, cdBasename)
	return nil
}
```

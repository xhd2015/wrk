# Scenario

**Feature**: fallback wrk --cd myrepo single projects match resolves saved abs

```
saved/myrepo recorded; no local ./myrepo; channel closed; fake bash
wrk --cd myrepo -> stdout saved abs; shell cwd = saved abs
```

## Steps

1. Record one saved project basename `myrepo`.
2. Neutral cwd; install fake bash; `wrk --cd myrepo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	saved := initSavedGitRepo(t, req.WorkRoot, "saved", cdBasename)
	recordSavedProject(t, req, saved)
	req.MainRepo = resolvePath(t, saved)
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	installFakeBash(t, req, 0)
	setCDFlagThenPath(req, cdBasename)
	return nil
}
```

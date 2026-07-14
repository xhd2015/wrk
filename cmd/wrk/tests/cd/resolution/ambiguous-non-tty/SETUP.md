# Scenario

**Feature**: wrk --cd myrepo with two saved matches on non-TTY errors

```
aaa/myrepo + zzz/myrepo saved; no local ./myrepo
non-TTY wrk --cd myrepo -> non-zero; lists both candidates
```

## Steps

1. Record two projects with basename `myrepo`.
2. Run `wrk --cd myrepo` without basename confirm env.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	repoA := initSavedGitRepo(t, req.WorkRoot, "aaa", cdBasename)
	repoZ := initSavedGitRepo(t, req.WorkRoot, "zzz", cdBasename)
	recordSavedProject(t, req, repoA)
	recordSavedProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	setCDFlagThenPath(req, cdBasename)
	return nil
}
```

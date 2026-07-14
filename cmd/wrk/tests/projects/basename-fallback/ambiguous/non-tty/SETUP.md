# Scenario

**Feature**: ambiguous basename in non-TTY errors with candidate list

```
aaa/myrepo + zzz/myrepo saved
non-TTY wrk myrepo -> error listing both absolute paths; no worktree
```

## Steps

1. Create and record `{WorkRoot}/aaa/myrepo` and `{WorkRoot}/zzz/myrepo`.
2. Run `wrk myrepo` from neutral cwd without `WRK_BASENAME_CONFIRM`.

```go
func Setup(t *testing.T, req *Request) error {
	repoA := initSavedGitRepo(t, req.WorkRoot, "aaa", "myrepo")
	repoZ := initSavedGitRepo(t, req.WorkRoot, "zzz", "myrepo")
	recordSavedProject(t, req, repoA)
	recordSavedProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	return nil
}
```
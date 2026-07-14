# Scenario

**Feature**: ambiguous basename resolved via TTY prompt with stdin selection for --status

```
aaa/myrepo + zzz/myrepo saved
WRK_BASENAME_CONFIRM=1 + stdin "2" -> status for zzz/myrepo (lex-sorted #2)
```

## Steps

1. Create git repos `{WorkRoot}/aaa/myrepo` and `{WorkRoot}/zzz/myrepo`.
2. Record both with `wrk --add`.
3. Run `wrk myrepo --status` from `{WorkRoot}/workspace` with `WRK_BASENAME_CONFIRM=1` and stdin `2\n`.

```go
func Setup(t *testing.T, req *Request) error {
	repoA := initSavedGitRepo(t, req.WorkRoot, "aaa", "myrepo")
	repoZ := initSavedGitRepo(t, req.WorkRoot, "zzz", "myrepo")
	recordSavedProject(t, req, repoA)
	recordSavedProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	req.SelectedSavedRepo = repoZ
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"
	req.StdinInput = "2\n"
	return nil
}
```

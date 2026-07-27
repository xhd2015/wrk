# Scenario

**Feature**: wrk --where spl with two saved matches prints both paths

```
aaa/spl + zzz/spl saved
workspace/ -> wrk --where spl -> both abs paths sorted, one per line
```

## Steps

1. Create and record `{WorkRoot}/aaa/spl` and `{WorkRoot}/zzz/spl`.
2. Run `wrk --where spl` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	repoA := initSavedGitRepo(t, req.WorkRoot, "aaa", whereBasename)
	repoZ := initSavedGitRepo(t, req.WorkRoot, "zzz", whereBasename)
	recordSavedProject(t, req, repoA)
	recordSavedProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = whereArgs(whereBasename)
	return nil
}```

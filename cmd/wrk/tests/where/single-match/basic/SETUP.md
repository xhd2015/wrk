# Scenario

**Feature**: wrk --where spl prints one saved absolute path

```
saved/spl in projects.json
workspace/ (cwd, no ./spl) -> wrk --where spl -> stdout saved abs path
```

## Steps

1. Create git repo at `{WorkRoot}/saved/spl`.
2. Record it with `wrk --add`.
3. Run `wrk --where spl` from neutral cwd `{WorkRoot}/workspace`.

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = whereArgs(whereBasename)
	return nil
}```

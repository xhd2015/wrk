# Scenario

**Feature**: wrk --where absolute path rejects non-basename input

```
saved/spl recorded at known absolute path
workspace/ -> wrk --where <abs>/saved/spl -> non-zero, basename-only rejection
```

## Steps

1. Create and record `{WorkRoot}/saved/spl`.
2. Run `wrk --where <absolute-path-to-saved/spl>` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = whereArgs(resolvePath(t, savedRepo))
	return nil
}```

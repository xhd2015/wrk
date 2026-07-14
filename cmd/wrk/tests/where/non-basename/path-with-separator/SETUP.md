# Scenario

**Feature**: wrk --where sub/spl rejects path separator

```
saved/spl recorded
workspace/ -> wrk --where sub/spl -> non-zero, basename-only rejection
```

## Steps

1. Create and record `{WorkRoot}/saved/spl`.
2. Run `wrk --where sub/spl` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = whereArgs("sub/" + whereBasename)
	return nil
}```

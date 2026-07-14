# Scenario

**Feature**: wrk --where spl ignores cwd ./spl non-git directory

```
WorkRoot/spl (non-git dir exists)
saved/spl recorded in projects.json
WorkRoot -> wrk --where spl -> stdout = saved path only
```

## Steps

1. Create non-git directory `{WorkRoot}/spl`.
2. Create and record saved git repo at `{WorkRoot}/saved/spl`.
3. Run `wrk --where spl` from `{WorkRoot}`.

```go
func Setup(t *testing.T, req *Request) error {
	localSpl := initNonGitBasenameInDir(t, req.WorkRoot, whereBasename)
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = req.WorkRoot
	req.Args = whereArgs(whereBasename)

	// Guard: local cwd entry must differ from saved project path.
	if resolvePath(t, localSpl) == resolvePath(t, savedRepo) {
		t.Fatalf("local and saved paths must differ for no-local-fallback scenario")
	}
	return nil
}```

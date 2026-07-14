# Scenario

**Feature**: wrk myrepo --repos lists saved repo paths from single projects.json match

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo --repos -> "."
```

## Steps

1. Create git repo at `{WorkRoot}/saved/myrepo`.
2. Record it with `wrk --add`.
3. Run `wrk myrepo --repos` from neutral cwd `{WorkRoot}/workspace`.

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	return nil
}
```
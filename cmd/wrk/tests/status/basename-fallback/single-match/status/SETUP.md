# Scenario

**Feature**: wrk myrepo --status reports saved repo from single projects.json match

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo --status -> one clean block for saved root
```

## Steps

1. Create git repo at `{WorkRoot}/saved/myrepo`.
2. Record it with `wrk --add`.
3. Run `wrk myrepo --status` from neutral cwd `{WorkRoot}/workspace`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	return nil
}
```

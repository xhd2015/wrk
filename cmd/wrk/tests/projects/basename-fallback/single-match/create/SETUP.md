# Scenario

**Feature**: wrk myrepo creates worktree from single saved project when cwd elsewhere

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo -> worktree from saved/myrepo
```

## Steps

1. Create git repo at `{WorkRoot}/saved/myrepo`.
2. Record it with `wrk --add`.
3. Run `wrk myrepo` from neutral cwd `{WorkRoot}/workspace`.

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
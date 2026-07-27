# Scenario

**Feature**: wrk myrepo --list prints saved repo worktree list from single projects.json match

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo --list -> git worktree list for saved root
```

## Steps

1. Create git repo at `{WorkRoot}/saved/myrepo`.
2. Record it with `wrk --add`.
3. Run `wrk myrepo --list` from neutral cwd `{WorkRoot}/workspace`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	req.Args = []string{"--list"}
	return nil
}
```
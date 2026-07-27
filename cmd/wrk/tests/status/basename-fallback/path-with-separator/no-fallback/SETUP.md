# Scenario

**Feature**: wrk saved/myrepo --status does not fall back to saved project basename myrepo

```
saved/myrepo recorded; workspace/ cwd -> wrk saved/myrepo --status -> does not exist (no fallback)
```

## Steps

1. Create and record saved git repo `{WorkRoot}/saved/myrepo`.
2. Run `wrk saved/myrepo --status` from neutral cwd (no local `saved/myrepo`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "saved/myrepo"
	return nil
}
```

# Scenario

**Feature**: wrk myrepo --done does not fall back to saved project

```
saved/myrepo recorded
workspace/ cwd -> wrk myrepo --done -> does not exist (no fallback)
```

## Steps

1. Create and record `{WorkRoot}/saved/myrepo`.
2. Run `wrk myrepo --done` from neutral cwd without local `./myrepo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "myrepo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "myrepo"
	req.Args = []string{"--done"}
	return nil
}
```
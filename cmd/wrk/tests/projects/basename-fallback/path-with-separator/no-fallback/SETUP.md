# Scenario

**Feature**: wrk sub/foo does not fall back to saved project basename foo

```
saved/foo recorded; workspace/ cwd -> wrk sub/foo -> does not exist (no fallback)
```

## Steps

1. Create and record saved git repo `{WorkRoot}/saved/foo`.
2. Run `wrk sub/foo` from neutral cwd (no local `sub/foo`).

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", "foo")
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.TargetDir = "sub/foo"
	return nil
}
```
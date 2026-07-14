# Scenario

**Feature**: wrk --where spl --status is mutually exclusive

```
saved/spl recorded
workspace/ -> wrk --where spl --status -> non-zero, mutually exclusive
```

## Steps

1. Create and record `{WorkRoot}/saved/spl`.
2. Run `wrk --where spl --status` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	savedRepo := initSavedGitRepo(t, req.WorkRoot, "saved", whereBasename)
	recordSavedProject(t, req, savedRepo)
	req.MainRepo = savedRepo
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--where", whereBasename, "--status"}
	return nil
}```

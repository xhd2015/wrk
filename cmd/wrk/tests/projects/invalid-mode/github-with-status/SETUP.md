# Scenario

**Feature**: --github with --status is rejected

```
wrk --status --github -> non-zero, stderr only valid with --projects
```

## Steps

1. Initialize a git repo.
2. Run `wrk --status --github`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.RepoDir = mainRepo
	req.Args = []string{"--status", "--github"}
	return nil
}
```

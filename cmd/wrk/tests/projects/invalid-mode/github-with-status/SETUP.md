# Scenario

**Feature**: --github with --status is rejected

```
wrk --status --github -> non-zero, stderr only valid with --projects
```

## Steps

1. Initialize a git repo.
2. Run `wrk --status --github`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.RepoDir = mainRepo
	req.Args = []string{"--status", "--github"}
	return nil
}
```

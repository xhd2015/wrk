# Scenario

**Feature**: --github without --projects is rejected

```
wrk --github -> non-zero, stderr only valid with --projects
```

## Steps

1. Initialize a git repo for a valid cwd.
2. Run `wrk --github`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.RepoDir = mainRepo
	req.Args = []string{"--github"}
	return nil
}
```

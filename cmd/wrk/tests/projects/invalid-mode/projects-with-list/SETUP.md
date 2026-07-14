# Scenario

**Feature**: wrk --projects --list is mutually exclusive

```
wrk --projects --list -> non-zero exit, stderr mentions mutual exclusion
```

## Steps

1. Initialize a git repo (so cwd could otherwise be valid).
2. Run `wrk --projects --list`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.RepoDir = mainRepo
	req.Args = []string{"--projects", "--list"}
	return nil
}
```
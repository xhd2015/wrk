# Scenario

**Feature**: wrk --rm with --list is mutually exclusive

```
wrk --rm X --list -> non-zero exit, stderr mentions mutual exclusion
```

## Steps

1. Initialize a git repo (so cwd could otherwise be valid).
2. Run `wrk --rm <mainRepo> --list`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--rm", mainRepo, "--list"}
	return nil
}
```

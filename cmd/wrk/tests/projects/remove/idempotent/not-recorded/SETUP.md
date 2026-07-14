# Scenario

**Feature**: wrk --rm on never-recorded path is a no-op

```
myrepo (never added) -> wrk --rm myrepo -> exit 0, empty stdout, no projects.json entry
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` (do not record).
2. Run `wrk --rm <mainRepo>`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--rm", mainRepo}
	return nil
}
```

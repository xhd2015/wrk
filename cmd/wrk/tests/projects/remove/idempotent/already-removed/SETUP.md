# Scenario

**Feature**: second wrk --rm on same path is idempotent

```
wrk --add -> wrk --rm -> wrk --rm again -> exit 0, empty stdout
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --add <mainRepo>`.
3. Run `wrk --rm <mainRepo>` (first remove).
4. Run `wrk --rm <mainRepo>` (test invocation).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	recordProjectViaAdd(t, req, mainRepo)
	runWrkWithArgs(t, req, req.WorkRoot, "--rm", mainRepo)
	req.MainRepo = mainRepo
	req.Args = []string{"--rm", mainRepo}
	return nil
}
```

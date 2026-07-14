# Scenario

**Feature**: wrk --rm from main repo path

```
wrk --add myrepo -> wrk --rm myrepo -> stdout main path; entry gone from projects.json
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --add <mainRepo>` to record.
3. Run `wrk --rm <mainRepo>`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	recordProjectViaAdd(t, req, mainRepo)
	req.MainRepo = mainRepo
	req.Args = []string{"--rm", mainRepo}
	return nil
}
```

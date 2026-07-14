# Scenario

**Feature**: wrk --add from main repo path

```
wrk --add myrepo -> stdout myrepo main path; projects.json source manual
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --add <mainRepo>`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.Args = []string{"--add", mainRepo}
	return nil
}
```
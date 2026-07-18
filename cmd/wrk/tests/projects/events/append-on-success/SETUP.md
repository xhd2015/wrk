# Scenario

**Feature**: successful wrk create via `--new` appends event with exit_code 0

```
myrepo -> wrk --new (create) -> events.jsonl line with command create, exit_code 0
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --new` (P1 create entry) from main repo.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--new"}
	return nil
}
```

# Scenario

**Feature**: failed wrk invocation still appends event with non-zero exit_code

```
myrepo (main cwd) -> wrk --done -> error (not linked wt); event exit_code != 0
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --done` from main repo (not a linked worktree).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--done"}
	return nil
}
```
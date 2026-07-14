# Scenario

**Feature**: wrk -v on no-args create logs and streams worktree add

```
no-args create -> stderr has timestamp worktree add log + git subprocess lines
minor reads (rev-parse, status) not logged
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
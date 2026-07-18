# Scenario

**Feature**: wrk -v on create via --new logs and streams worktree add

```
create via --new -> stderr has timestamp worktree add log + git subprocess lines
minor reads (rev-parse, status) not logged
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
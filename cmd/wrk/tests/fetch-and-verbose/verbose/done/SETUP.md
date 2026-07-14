# Scenario

**Feature**: wrk --done -v logs merge-back major git commands

```
linked wt already-included -> wrk --done -v -> stderr merge and/or worktree remove
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
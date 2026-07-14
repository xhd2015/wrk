# Scenario

**Feature**: wrk --list -v does not log minor git worktree list

```
worktree list is read-only introspection -> stderr empty
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
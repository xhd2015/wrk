# Scenario

**Feature**: wrk --list -v does not log minor git worktree list

```
worktree list is read-only introspection -> stderr empty
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
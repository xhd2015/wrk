# Scenario

**Feature**: `-y` auto-confirms own-worktree `wrk --merge-back` prompts

```
wrk --merge-back -y -> merge without prompt; worktree kept
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

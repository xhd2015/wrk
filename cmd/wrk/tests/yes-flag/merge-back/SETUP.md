# Scenario

**Feature**: `-y` auto-confirms own-worktree `wrk --merge-back` prompts

```
wrk --merge-back -y -> merge without prompt; worktree kept
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

# Scenario

**Feature**: `-y` auto-confirms own-worktree `wrk --done` merge-back prompts

```
# consumer linked wt ahead of main; -y skips Proceed? and completes merge-back --rm
wrk --done -y -> ff-merge + remove (no stdin read on non-TTY)
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

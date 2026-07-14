# Scenario

**Feature**: `-y` cascade behavior — blocked on non-TTY, auto-yes on TTY

```
# non-TTY: ahead external dep -> wrk --done -y -> error (cascade guard)
# TTY: ahead external dep -> wrk --done -y -> cascade + consumer merge-back succeed
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

# Scenario

**Feature**: `-y` cascade behavior — auto-yes on TTY and non-TTY

```
# non-TTY: ahead external dep -> wrk --done -y -> cascade + consumer succeed
# TTY: ahead external dep -> wrk --done -y -> cascade + consumer merge-back succeed
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

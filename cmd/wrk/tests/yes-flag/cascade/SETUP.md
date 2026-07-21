# Scenario

**Feature**: `-y` cascade behavior — synonym of default auto-yes on TTY and non-TTY

```
# non-TTY: ahead external dep -> wrk --done -y -> exit 0 cascade + consumer
# TTY: ahead external dep -> wrk --done -y -> cascade + consumer merge-back succeed
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

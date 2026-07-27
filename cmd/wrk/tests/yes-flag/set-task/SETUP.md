# Scenario

**Feature**: `-y` auto-confirms `wrk --set-task` rename prompt

```
wrk --set-task "new" -y -> rename without stdout-TTY check
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

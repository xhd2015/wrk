# Scenario

**Feature**: `-y` auto-confirms `wrk --set-task` rename prompt

```
wrk --set-task "new" -y -> rename without stdout-TTY check
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

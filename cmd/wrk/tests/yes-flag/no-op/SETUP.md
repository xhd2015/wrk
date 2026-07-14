# Scenario

**Feature**: `-y` is a no-op on commands without Y/n prompts

```
wrk -y (create) -> same behavior as bare wrk
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

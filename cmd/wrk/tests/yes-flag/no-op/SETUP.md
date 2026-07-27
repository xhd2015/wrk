# Scenario

**Feature**: `-y` is a no-op on commands without Y/n prompts

```
wrk -y (create) -> same behavior as bare wrk
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

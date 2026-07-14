# Scenario

**Feature**: create with no window/terminal/agent axes

```
empty config + bare wrk -> native create only; no space/iterm/agent
```

## Steps

- Leaves use empty config and no UX flags.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

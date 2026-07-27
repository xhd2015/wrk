# Scenario

**Feature**: create UX pipeline success paths

```
create -> [window] -> [terminal ± agent follow-up | agent-in-process]
```

## Steps

- Subtrees split flags-driven vs config-driven effective UX.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

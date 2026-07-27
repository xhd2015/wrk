# Scenario

**Feature**: second positional is not task-like — unchanged target-dir create

```
wrk <dir> ./out | out | <target> with -t already set
  -> no treat-as-task prompt; normal target-dir semantics
```

## Steps

- Leaves configure path-like, short token, or explicit `-t` + second positional.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping node: leaves configure positionals and interactive mode.
	skipIfNoGit(t)
	return nil
}
```

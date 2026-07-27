# Scenario

**Feature**: one positional `wrk <arg1>` — task-like promote from cwd vs normal source resolve

```
wrk <arg1>
  -> if task-like and does not resolve as source: confirm / -y / non-TTY error
  -> if existing source path/basename: normal create; no prompt
```

## Steps

- Leaves use `setupOneArg` (cwd = mainRepo) or custom source-resolve setups.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping node: leaves configure positionals and interactive mode.
	skipIfNoGit(t)
	return nil
}
```

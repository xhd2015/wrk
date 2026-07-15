# Scenario

**Feature**: one-arg task-like text that does not resolve as a source directory

```
cwd=git-repo; wrk "fix the login bug"
  -> treat as --task (create from cwd) or non-TTY error+hint
```

## Steps

- Descendants set interactive mode.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping node: leaves configure positionals and interactive mode.
	skipIfNoGit(t)
	return nil
}
```

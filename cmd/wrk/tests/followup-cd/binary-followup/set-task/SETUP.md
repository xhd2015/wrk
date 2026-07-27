# Scenario

**Feature**: --set-task follow-up only when path actually moves

```
wrk --set-task <desc> + WRK_FOLLOWUP_FILE
  -> move: cd <newPath>; unchanged: empty

# --force-cd bypasses cwd-missing gate from surviving sibling
sibling A (cwd); rename B --force-cd + env -> cd <newPath>
sibling A; rename B --force-cd (no channel) -> shell @ newPath
```

## Steps

1. Descendants create task worktree and set SetTaskDesc / CLIArgs.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "binary")
	return nil
}
```

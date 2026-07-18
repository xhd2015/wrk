# Scenario

**Feature**: non-TTY bare `wrk` does not create a worktree

```
myrepo -> wrk (no args, non-TTY)
  -> no {WRK_HOME}/worktrees/myrepo-main-2026-06-30
  -> worktrees/ stays empty / absent
```

## Steps

1. Init main repo (parent).
2. Run bare `wrk` (Args empty; harness is non-TTY).
3. Assert create did not run.

```go
func Setup(t *testing.T, req *Request) error {
	// Parent sets main repo; force empty argv so mode is dashboard (not create).
	req.Args = nil
	req.TargetDir = ""
	req.TaskDesc = ""
	return nil
}
```

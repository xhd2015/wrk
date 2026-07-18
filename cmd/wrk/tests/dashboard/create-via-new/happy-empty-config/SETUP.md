# Scenario

**Feature**: `wrk --new` creates a worktree like old bare `wrk` (empty config)

```
myrepo (main) + empty config
  -> wrk --new
  -> exit 0; stdout {WRK_HOME}/worktrees/myrepo-main-2026-06-30\n
  -> linked worktree exists
```

## Steps

1. Init `myrepo` on main (parent).
2. Run `wrk --new` with no other flags.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--new"}
	return nil
}
```

# Scenario

**Feature**: P3 interactive dashboard — CANCEL + RUN real compose (hermetic env, no PTY)

```
# Force action path without TTY
WRK_DASHBOARD_ACTION=cancel|run-done|run-merge-back
  (+ WRK_DASHBOARD_DRY_RUN=1 for plan-only RUN)
  (+ WRK_DASHBOARD_COMPOSE_ARGV_LOG for recipe assert)
  -> bare wrk honors action even when stdin is non-TTY

# cancel
  -> exit 0; no compose mutations; event command=dashboard

# run-done (defaults)
  -> real compose ≡
     wrk --gen-commit-msg [--add-all] --commit --agent-runner=commandcode
         --done --sync --tag-next --push --reinstall-local [--dry-run]

# run-merge-back
  -> same with --merge-back instead of --done

# gates
  -> disabled add changes cannot be forced on → no --add-all

# non-TTY without ACTION
  -> P2 static snapshot still (regression)
```

## Preconditions

- See parent `dashboard/SETUP.md` **P3 hermetic env contract**.
- Linked worktree fixtures use `git worktree add` (not under `WRK_HOME/worktrees`).

## Steps

- Leaves set `ExtraEnv` via `setDashboardAction` / `setDashboardToggles` and empty `Args`.
- RUN leaves use `WRK_DASHBOARD_DRY_RUN=1` for hermetic zero-mutation compose.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.Args = nil
	req.TargetDir = ""
	req.TaskDesc = ""
	return nil
}
```

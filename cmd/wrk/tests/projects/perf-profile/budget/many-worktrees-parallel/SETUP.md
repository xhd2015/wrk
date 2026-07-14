# Scenario

```
# serial today: 12 worktrees -> worktree_status_all ~190ms+, run_end ~300ms+
# parallel target: worktree_status_all < 100ms, run_end < 200ms
main repo + 12 linked worktrees -> wrk --projects perf budget
```

## Steps

1. Same fixture as emits-events (12 clean linked worktrees, no upstream fetch noise).
2. Assert perf log budgets encoding the parallel-gather fix.

```go
func Setup(t *testing.T, req *Request) error {
	ensurePerfProfileHelpersUsed()
	setupPerfProfileRepo(t, req, "perf-budget", 12)
	return nil
}
```
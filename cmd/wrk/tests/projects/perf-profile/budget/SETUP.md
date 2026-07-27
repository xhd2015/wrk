# Scenario

**Feature**: wrk --projects performance budgets (parallel gather)

```
N linked worktrees -> worktree_status_all and run_end under latency ceilings
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensurePerfProfileHelpersUsed()
	return nil
}
```
# Scenario

```
# today: list_linked_skip + list_linked_summary (duplicate ListLinked per project)
# target: single ListLinked shared by main status + worktree summary
main repo + 3 linked worktrees -> exactly one list_linked phase in perf log
```

## Steps

1. Create main repo with 3 linked worktrees.
2. Assert perf log records only one `list_linked*` phase per project.

```go
func Setup(t *testing.T, req *Request) error {
	ensurePerfProfileHelpersUsed()
	setupPerfProfileRepo(t, req, "perf-dedup", 3)
	return nil
}
```
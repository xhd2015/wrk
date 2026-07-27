# Scenario

```
main repo + 12 linked worktrees + WRK_PROJECTS_PERF_LOG -> JSONL with run/project/phase/worktree events
```

## Steps

1. Create tracked main repo with 12 linked worktrees.
2. Record project and run `wrk --projects` with perf log enabled.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensurePerfProfileHelpersUsed()
	setupPerfProfileRepo(t, req, "perf-emit", 12)
	return nil
}
```
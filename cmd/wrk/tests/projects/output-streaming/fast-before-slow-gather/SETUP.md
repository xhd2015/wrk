# Scenario

**Feature**: wrk --projects streams the fast lex-first project before slow project gather completes

```
aaa (clean, no upstream, no worktrees) + zzz (12 linked worktrees, tracked origin)
-> wrk --projects -> first stdout within ~50ms; total run >> first byte (zzz still gathering)
```

## Steps

1. Create `aaa` (minimal main repo, no remote).
2. Create `zzz` with bare `origin`, 12 linked worktrees.
3. Record both via `wrk --add`.
4. Run `wrk --projects` (doctest Run + streaming probe in Assert).

```go
func Setup(t *testing.T, req *Request) error {
	ensureOutputStreamingHelpersUsed()

	repoA := setupFastNoUpstreamRepo(t, req.WorkRoot, "aaa")
	repoZ := setupSlowManyWorktreesRepo(t, req, "zzz", 12)
	recordStreamingProject(t, req, repoA)
	recordStreamingProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	return nil
}
```
# Scenario

**Feature**: fixed-path create with pre-existing preferred branch suffixes branch only (P0 C3)

```
# parent exists, fixed <target-dir> free; preferred branch main-{date} already exists
myrepo (main) + refs/heads/main-2026-06-30
  -> wrk myrepo {WorkRoot}/wt
  -> path stays {WorkRoot}/wt; branch main-2026-06-30-1 via worktree add -b (no reuse / no --no-checkout)
```

## Steps

1. Parent setup sets `req.SpawnDir = {WorkRoot}/wt`.
2. Pre-create branch `main-2026-06-30` (ref only; no worktree at the fixed path).
3. Run `wrk myrepo {WorkRoot}/wt` from process cwd `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	runGitIsolated(t, req.TargetDir, "branch", branchName("main", wrkDate, 0))
	return nil
}
```

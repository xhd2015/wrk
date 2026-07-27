# Scenario

**Feature**: after successful `--done -y --tag-next`, create local root-bump tag at main HEAD

```
# wt ahead of v0.0.1 baseline → done merges; tag-next creates v0.0.2 at main HEAD
myrepo (v0.0.1) + wt (feature-work)
  -> wrk --done -y --tag-next
  -> merged branch <WtBranch> into main
  -> <blank>
  -> plan/apply: v0.0.1 owned changed -> v0.0.2; tagged; 1 tag created
  -> local v0.0.2 @ main HEAD; no push required
  -> event command "done"
```

## Steps

1. Seed main-gomod; tag `v0.0.1`; create wrk-managed linked worktree; commit ahead.
2. Run `wrk --done -y --tag-next` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineLocal(t, req)
	req.Args = []string{"--done", "-y", "--tag-next"}
	return nil
}
```

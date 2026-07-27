# Scenario

**Feature**: after successful `--done -y --push`, main tip is on bare origin

```
# wt ahead; bare origin has pre-merge tip → done merges + pushes main
myrepo (origin) + wt (feature-work)
  -> wrk --done -y --push
  -> merged branch <WtBranch> into main
  -> <blank>
  -> pushed main → origin/main
  -> wt gone; origin/main == main HEAD; no tags required
```

## Steps

1. Seed main repo; attach bare `origin` and `push -u origin main`.
2. Create wrk-managed linked worktree; commit ahead (`feature-work`).
3. Run `wrk --done -y --push` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePushWithOrigin(t, req)
	req.Args = []string{"--done", "-y", "--push"}
	return nil
}
```

# Scenario

**Feature**: `--done -y --push` with no resolvable remote fails clearly after successful merge

```
# no origin / no upstream on main → after done, push helper cannot resolve remote
myrepo (no remote) + wt (feature-work)
  -> wrk --done -y --push
  -> primary merge-back may succeed (wt removed)
  -> non-zero exit; stderr explains missing remote
```

## Steps

1. Create wrk-managed linked worktree and commit ahead (no remote).
2. Run `wrk --done -y --push` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePushNoRemote(t, req)
	req.Args = []string{"--done", "-y", "--push"}
	return nil
}
```

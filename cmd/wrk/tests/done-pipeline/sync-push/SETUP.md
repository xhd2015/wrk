# Scenario

**Feature**: `--done -y --sync --push` (no `--tag-next`) labels ship lane `push`, not `tag-next+push`

```
# wtA ahead; wtB behind; origin present
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --done -y --sync --push
  -> merge → concurrent ship: push ‖ sync
  -> stderr progress label "push" (not "tag-next+push"); no new tag
  -> wtA gone; origin/main == main
```

## Steps

1. Root-bump seed + bare origin + two worktrees; commit ahead on wtA.
2. Run `wrk --done -y --sync --push` from wtA.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineSyncWithOrigin(t, req)
	req.Args = []string{"--done", "-y", "--sync", "--push"}
	return nil
}
```

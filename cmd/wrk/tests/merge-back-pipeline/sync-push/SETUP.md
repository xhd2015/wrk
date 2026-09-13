# Scenario

**Feature**: `--merge-back -y --sync --push` (no `--tag-next`) labels ship lane `push`, not `tag-next+push`

```
# wtA ahead; wtB behind; origin present
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --merge-back -y --sync --push
  -> merge → concurrent ship: push ‖ sync
  -> stderr progress label "push" (not "tag-next+push"); no new tag
  -> wtA kept; origin/main == main
```

## Steps

1. Root-bump seed + bare origin + two worktrees; commit ahead on wtA.
2. Run `wrk --merge-back -y --sync --push` from wtA.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMergeBackPipelineSyncWithOrigin(t, req)
	req.Args = []string{"--merge-back", "-y", "--sync", "--push"}
	return nil
}
```

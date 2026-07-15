# Scenario

**Feature**: full post-pipeline `--done -y --sync --tag-next --push` ordered stdout + side effects

```
# wtA ahead; wtB behind; origin present
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --done -y --sync --tag-next --push
  -> merge → blank → sync → blank → tag-next → blank → push
  -> wtA gone; wtB updated; local+origin v0.0.2; origin/main == main
```

## Steps

1. Root-bump seed + bare origin + two worktrees; commit ahead on wtA.
2. Run `wrk --done -y --sync --tag-next --push` from wtA.

```go
func Setup(t *testing.T, req *Request) error {
	setupDonePipelineSyncWithOrigin(t, req)
	req.Args = []string{"--done", "-y", "--sync", "--tag-next", "--push"}
	return nil
}
```

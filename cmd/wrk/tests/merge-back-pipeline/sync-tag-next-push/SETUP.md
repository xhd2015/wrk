# Scenario

**Feature**: full post-pipeline `--merge-back -y --sync --tag-next --push` ordered stdout + side effects; wt kept

```
# wtA ahead; wtB behind; origin present
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --merge-back -y --sync --tag-next --push
  -> merge → blank → sync → blank → tag-next → blank → push
  -> wtA remains; wtB updated; local+origin v0.0.2; origin/main == main
```

## Steps

1. Root-bump seed + bare origin + two worktrees; commit ahead on wtA.
2. Run `wrk --merge-back -y --sync --tag-next --push` from wtA.

```go
func Setup(t *testing.T, req *Request) error {
	setupMergeBackPipelineSyncWithOrigin(t, req)
	req.Args = []string{"--merge-back", "-y", "--sync", "--tag-next", "--push"}
	return nil
}
```

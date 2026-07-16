# Scenario

**Feature**: `--tag-next --push` creates tags then pushes **branch + tags** (runPushMain semantics)

```
# root bump + origin remote
git repo + bare origin
  -> wrk --tag-next --push
  -> local tag v0.0.2
  -> blank line after tag-next human block
  -> pushed main → origin/main
  -> origin has refs/tags/v0.0.2 AND refs/heads/main == local HEAD
```

## Steps

1. `setupPushRepo` (repo with origin bare remote).
2. Run `wrk --tag-next --push`.

```go
func Setup(t *testing.T, req *Request) error {
	setupPushRepo(t, req)
	// Advance local main after origin was seeded so tags-only push cannot
	// accidentally satisfy "branch tip on origin" (origin lags local HEAD).
	commitFile(t, req.MainRepo, "README.md", "# v3 post-origin\n", "local tip after origin seed")
	req.Args = []string{"--tag-next", "--push"}
	return nil
}
```

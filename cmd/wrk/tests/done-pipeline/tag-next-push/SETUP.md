# Scenario

**Feature**: `--done -y --tag-next --push` creates local tags then pushes branch + tags to origin

```
# origin + root-bump seed; after done, tag-next then runPushMain(main, createdTags)
myrepo (origin, v0.0.1) + wt (feature-work)
  -> wrk --done -y --tag-next --push
  -> merged branch <WtBranch> into main
  -> <blank>
  -> tag-next apply (v0.0.2 local)
  -> <blank>
  -> pushed main → origin/main
  -> origin/main == main HEAD; origin has refs/tags/v0.0.2
```

## Steps

1. Seed main-gomod; tag `v0.0.1`; attach bare origin; create wrk wt; commit ahead.
2. Run `wrk --done -y --tag-next --push` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineWithOrigin(t, req)
	req.Args = []string{"--done", "-y", "--tag-next", "--push"}
	return nil
}
```

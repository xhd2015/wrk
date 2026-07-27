# Scenario

**Feature**: `--merge-back -y --tag-next --push` creates local tags then pushes branch + tags; worktree kept

```
# origin + root-bump seed; after merge-back, tag-next then runPushMain(main, createdTags)
myrepo (origin, v0.0.1) + wt (feature-work)
  -> wrk --merge-back -y --tag-next --push
  -> merged branch <WtBranch> into main
  -> <blank>
  -> tag-next apply (v0.0.2 local)
  -> <blank>
  -> pushed main → origin/main
  -> origin/main == main HEAD; origin has refs/tags/v0.0.2
  -> source wt remains
```

## Steps

1. Seed main-gomod; tag `v0.0.1`; attach bare origin; create wrk wt; commit ahead.
2. Run `wrk --merge-back -y --tag-next --push` from the worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMergeBackPipelineWithOrigin(t, req)
	req.Args = []string{"--merge-back", "-y", "--tag-next", "--push"}
	return nil
}
```

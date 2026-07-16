# Scenario

**Feature**: `--merge-back -y --tag-next --propagate-tags` creates next tag then bumps consumer; source wt kept

```
# lib wt ahead; app requires v0.0.1
linked wt + registered app
  -> wrk --merge-back -y --tag-next --propagate-tags
  -> primary merge (no remove)
  -> blank → tag-next v0.0.2
  -> blank → propagate app bump+commit
  -> source wt remains; event command "merge-back"
```

## Steps

1. Multi-project fixture (`setupMergeBackPipelinePropagateTagNext`).
2. Run `wrk --merge-back -y --tag-next --propagate-tags` from linked worktree.

```go
func Setup(t *testing.T, req *Request) error {
	setupMergeBackPipelinePropagateTagNext(t, req)
	req.Args = []string{"--merge-back", "-y", "--tag-next", "--propagate-tags"}
	return nil
}
```

# Scenario

**Feature**: aborted `--done` runs no sync / tag-next / push / propagate even when all flags present

```
# user declines confirm → merge-back aborted; no post-pipeline
myrepo (v0.0.1) + wtA (ahead)
  -> wrk --done --confirm-from-stdin --sync --tag-next --push --propagate-tags  (stdin: n)
  -> merge-back aborted
  -> no synced:; no tag created; no pushed line; no propagate stage; wt remains
```

## Steps

1. Root-bump seed + linked worktree ahead (origin optional; not required for abort).
2. Run with full post flags including `--propagate-tags` and stdin `n\n`.
   (No multi-project fixture required: abort stops before post stages.)

```go
func Setup(t *testing.T, req *Request) error {
	setupDonePipelineLocal(t, req)
	req.Args = []string{"--done", "--confirm-from-stdin", "--sync", "--tag-next", "--push", "--propagate-tags"}
	req.StdinInput = "n\n"
	return nil
}
```

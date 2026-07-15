# Scenario

**Feature**: aborted `--merge-back` runs no sync / tag-next / push even when all flags present

```
# user declines confirm → merge-back aborted; no post-pipeline
myrepo (v0.0.1) + wtA (ahead)
  -> wrk --merge-back --confirm-from-stdin --sync --tag-next --push  (stdin: n)
  -> merge-back aborted
  -> no synced:; no tag created; no pushed line; wt remains
```

## Steps

1. Root-bump seed + linked worktree ahead (origin optional; not required for abort).
2. Run with full post flags and stdin `n\n`.

```go
func Setup(t *testing.T, req *Request) error {
	setupMergeBackPipelineLocal(t, req)
	req.Args = []string{"--merge-back", "--confirm-from-stdin", "--sync", "--tag-next", "--push"}
	req.StdinInput = "n\n"
	return nil
}
```

# Scenario

**Feature**: aborted `--done` runs no sync / tag-next / push / propagate / reinstall even when all flags present

```
# user declines via --confirm → merge-back aborted; no post-pipeline / reinstall tail
myrepo (v0.0.1) + wtA (ahead)
  -> wrk --done --confirm --confirm-from-stdin --sync --tag-next --push --propagate-tags --reinstall-local  (stdin: n)
  -> merge-back aborted
  -> no synced:; no tag created; no pushed line; no propagate stage; no reinstall; wt remains
```

## Steps

1. Root-bump seed + linked worktree ahead (origin optional; not required for abort).
2. Run with full post flags including `--propagate-tags` and `--reinstall-local`, stdin `n\n`.
   (No multi-project / GOBIN fixture required: abort stops before post stages.)

```go
func Setup(t *testing.T, req *Request) error {
	setupDonePipelineLocal(t, req)
	req.Args = []string{"--done", "--confirm", "--confirm-from-stdin", "--sync", "--tag-next", "--push", "--propagate-tags", "--reinstall-local"}
	req.StdinInput = "n\n"
	return nil
}
```

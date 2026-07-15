# Scenario

**Feature**: `--merge-back --tag-next --dry-run` plans merge (keep wt) + tag-next; no mutations

```
# wt ahead of v0.0.1 → dry-run plans ff-merge + tag v0.0.2; wt stays; no tag created
myrepo (v0.0.1) + wt (feature-work)
  -> wrk --merge-back --tag-next --dry-run
  -> planned merge --ff-only (no remove)
  -> blank → 1 tag planned (would-be main tip)
  -> wt remains; no v0.0.2
```

## Steps

1. Root-bump seed + linked worktree ahead.
2. Snapshot baseline.
3. Run `wrk --merge-back --tag-next --dry-run` without `-y`.

```go
func Setup(t *testing.T, req *Request) error {
	setupMergeBackPipelineLocal(t, req)
	recordMergeBackDryRunBaseline(t, req)
	req.Args = []string{"--merge-back", "--tag-next", "--dry-run"}
	return nil
}
```

# Scenario

**Feature**: bare `wrk --pr` exits 0 with empty stdout when no open PR exists for the head branch

```
# remote feature exists; fake gh list → []
linked wt + origin/feature-pr + empty open-PR list
  -> wrk --pr
  -> exit 0; empty stdout (no warning required)
  -> no gh pr create / comment; list may run; no push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh (default list JSON `[]`).
3. Run bare `--pr` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	// Default FAKE_GH_LIST_JSON is [] — no setFakeGhExistingPR.
	req.Args = prShowArgs()
	return nil
}
```

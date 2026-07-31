# Scenario

**Feature**: no open PR for head → `--pr --status` exit 0, empty stdout, stderr warning

```
# remote feature exists; fake gh list → []
linked wt + origin/feature-pr + empty open-PR list
  -> wrk --pr --status
  -> exit 0; stdout empty
  -> stderr contains warning: and no open / pull request (prefer naming branch)
  -> no gh pr view / create / comment; no push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh (default list JSON `[]`).
3. Run `--pr --status` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	// Default FAKE_GH_LIST_JSON is [] — no setFakeGhExistingPR.
	req.Args = prStatusArgs()
	return nil
}
```

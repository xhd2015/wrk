# Scenario

**Feature**: open PR with in-progress checks → Checks=pending; exit 0

```
# remote feature exists; open PR; view rollup IN_PROGRESS / PENDING
linked wt + origin/feature-pr + fake gh list → [{number:42,…}]
  + FAKE_GH_VIEW_JSON statusCheckRollup PENDING
  -> wrk --pr --status
  -> stdout Checks: pending; exit 0
  -> no create/comment/push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; open PR list + pending rollup view JSON.
3. Run `--pr --status` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	setFakeGhViewJSON(t, req, prViewJSON(
		prExistingNumber, prExistingTitle, prDefaultURL, "OPEN", false,
		"REVIEW_REQUIRED", prRollupPendingJSON,
	))
	req.Args = prStatusArgs()
	return nil
}
```

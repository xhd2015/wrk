# Scenario

**Feature**: open PR with failing checks → `--status --pr` still exit 0; Checks=failure (report not gate)

```
# remote feature exists; open PR; view rollup FAILURE
linked wt + origin/feature-pr + fake gh list → [{number:42,…}]
  + FAKE_GH_VIEW_JSON statusCheckRollup FAILURE
  -> wrk --status --pr   # flag order free
  -> stdout Checks: failure; exit 0 (not a gate)
  -> no create/comment/push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; open PR list + failure rollup view JSON.
3. Run `--status --pr` (reverse flag order) from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	setFakeGhViewJSON(t, req, prViewJSON(
		prExistingNumber, prExistingTitle, prDefaultURL, "OPEN", false,
		"REVIEW_REQUIRED", prRollupFailureJSON,
	))
	// Flag order free: --status before --pr.
	req.Args = prStatusThenPrArgs()
	return nil
}
```

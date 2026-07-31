# Scenario

**Feature**: open PR with all checks green → `--pr --status` prints URL + State/Title/Checks=success/Reviews

```
# remote feature exists; open PR #42; view rollup SUCCESS
linked wt + origin/feature-pr + fake gh list → [{number:42,…}]
  + FAKE_GH_VIEW_JSON statusCheckRollup SUCCESS
  -> wrk --pr --status
  -> list open PR; gh pr view 42 --json …
  -> stdout: URL + State open + Title + Checks success + Reviews review required
  -> exit 0; no create/comment/push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; list JSON = open PR #42; view JSON = success rollup + REVIEW_REQUIRED.
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
		"REVIEW_REQUIRED", prRollupSuccessJSON,
	))
	req.Args = prStatusArgs()
	return nil
}
```

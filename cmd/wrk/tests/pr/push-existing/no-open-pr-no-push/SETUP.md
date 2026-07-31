# Scenario

**Feature**: push-existing with no open PR for head → non-zero; origin tip unchanged; no create

```
# origin/feature-pr stale; local ahead; fake gh list → []
linked wt (ahead) + origin/feature-pr (stale) + empty open-PR list
  -> wrk --pr --push
  -> list open PR FIRST
  -> non-zero; stderr mentions no open pull request (preferably head branch)
  -> origin tip UNCHANGED (no push); no gh pr create / comment
```

## Steps

1. Seed linked feature with remote present + local ahead (snapshot origin tip).
2. Install fake gh (default list JSON `[]`).
3. Run `--pr --push` without title/comment.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	// Default FAKE_GH_LIST_JSON is [] — no setFakeGhExistingPR.
	req.Args = prPushExistingArgs()
	return nil
}
```

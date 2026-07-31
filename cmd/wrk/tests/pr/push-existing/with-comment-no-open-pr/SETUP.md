# Scenario

**Feature**: push+comment with no open PR → non-zero; no push; no comment; no create

```
# origin/feature-pr stale; local ahead; fake gh list → []
linked wt (ahead) + empty open-PR list
  -> wrk --pr --push --comment "please review"
  -> list open PR FIRST
  -> non-zero; stderr mentions no open pull request
  -> origin tip UNCHANGED; no gh pr create / comment
```

## Steps

1. Seed linked feature with remote present + local ahead (snapshot origin tip).
2. Install fake gh (default list JSON `[]`).
3. Run `--pr --push --comment C` without `--title`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	// Default FAKE_GH_LIST_JSON is [] — no setFakeGhExistingPR.
	req.Args = prPushExistingCommentArgs(prDefaultComment)
	return nil
}
```

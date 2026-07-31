# Scenario

**Feature**: comment-only with no open PR for head → non-zero error; never create or comment

```
# remote feature exists; fake gh list → []
linked wt + origin/feature-pr + empty open-PR list
  -> wrk --pr --comment "please review"
  -> non-zero
  -> stderr mentions no open pull request (preferably head branch name)
  -> no gh pr create / comment; list may run; no push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh (default list JSON `[]`).
3. Run `--pr --comment C` without `--title`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	// Default FAKE_GH_LIST_JSON is [] — no setFakeGhExistingPR.
	req.Args = prCommentOnlyArgs(prDefaultComment)
	return nil
}
```

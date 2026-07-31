# Scenario

**Feature**: open PR for head → comment-only adds additive comment + prints URL; no create, no push, no title warning

```
# remote feature exists; open PR for feature-pr
linked wt + origin/feature-pr + fake gh list → [{number:42,url:…}]
  -> wrk --pr --comment "please review"   # no --title
  -> stdout: comment added\nURL
  -> stderr empty (no title-ignored warning — title was not passed)
  -> gh pr list + pr comment; pr create not called; no ensure-push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; set list JSON to one open PR (default URL/number).
3. Run `--pr --comment C` without `--title`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	req.Args = prCommentOnlyArgs(prDefaultComment)
	return nil
}
```

# Scenario

**Feature**: open PR → `--push --pr --comment C` full tip push then additive comment + URL

```
# origin/feature-pr stale; local ahead; open PR
linked wt (ahead) + origin/feature-pr (stale) + fake gh list → [{number:42,url:…}]
  -> wrk --push --pr --comment "please review"   # no --title; flag order free
  -> list open PR FIRST
  -> full push: origin tip advances
  -> gh pr comment; stdout: pushed …\n\ncomment added\nURL
  -> no gh pr create; no title-ignored warning
```

## Steps

1. Seed linked feature with remote present + local ahead (snapshot origin tip).
2. Install fake gh; set list JSON to one open PR.
3. Run `--push --pr --comment C` without `--title` (argv order free).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	// Prove flag order free: --push before --pr.
	req.Args = prPushThenPrCommentArgs(prDefaultComment)
	return nil
}
```

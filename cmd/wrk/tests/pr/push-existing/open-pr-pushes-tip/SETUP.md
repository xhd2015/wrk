# Scenario

**Feature**: open PR + local ahead → `--pr --push` full tip push + URL; no create, no comment

```
# origin/feature-pr stale; local one commit ahead; open PR for feature-pr
linked wt (ahead) + origin/feature-pr (stale) + fake gh list → [{number:42,url:…}]
  -> wrk --pr --push   # no --title / --comment
  -> list open PR FIRST
  -> full push: origin tip advances to local HEAD
  -> stdout: pushed feature-pr → origin/feature-pr\n\nURL
  -> no gh pr create / comment; no title-ignored warning
```

## Steps

1. Seed linked feature; push feature; commit local-ahead (snapshot origin tip).
2. Install fake gh; set list JSON to one open PR.
3. Run `--pr --push` without title/comment from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	req.Args = prPushExistingArgs()
	return nil
}
```

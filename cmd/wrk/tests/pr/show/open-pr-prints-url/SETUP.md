# Scenario

**Feature**: bare `wrk --pr` prints the open PR URL only when one exists for the head branch

```
# remote feature exists; open PR for feature-pr
linked wt + origin/feature-pr + fake gh list → [{url:…}]
  -> wrk --pr
  -> stdout: URL\n only (no PR created / comment added / title set / body set / pushed)
  -> gh pr list called; create/comment not called; no ensure-push
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; set list JSON to one open PR (default URL).
3. Run bare `--pr` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	req.Args = prShowArgs()
	return nil
}
```

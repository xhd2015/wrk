# Scenario

**Feature**: `--push --pr` with remote head present and local ahead updates origin tip then creates PR

```
# origin/feature-pr stale; local one commit ahead; no open PR
linked wt (ahead) + origin/feature-pr (stale) + fake gh list=[]
  -> wrk --push --pr --title "Fix login" --comment "please review"
  -> full push: origin tip advances to local HEAD
  -> stdout includes "pushed feature-pr → origin/feature-pr"
  -> then gh pr create (body = comment)
  -> PR created / title set / body set / URL
```

## Steps

1. Seed linked feature; push feature; commit local-ahead (snapshot origin tip).
2. Install fake gh (empty list → create path).
3. Run `--push --pr` with default title/comment from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	req.Args = prPushPrArgs()
	return nil
}
```

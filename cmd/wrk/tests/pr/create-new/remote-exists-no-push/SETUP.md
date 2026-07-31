# Scenario

**Feature**: when origin already has the head branch, `--pr` does not push (even if local is ahead)

```
# origin/feature-pr exists; local one commit ahead; no open PR
linked wt (ahead) + origin/feature-pr (stale tip)
  -> wrk --pr --title "Fix login" --comment "please review"
  -> no git push of tip; no "pushed …" line
  -> still gh pr create + comment
  -> PR created / title set / comment added / URL
  -> origin/feature-pr unchanged (still pre-ahead SHA)
```

## Steps

1. Seed linked feature; push feature to origin; commit local-ahead.
2. Snapshot origin tip; install fake gh (empty list).
3. Run bare `--pr` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```

# Scenario

**Feature**: existing open PR → title ignored warning + additive comment + URL; no push if remote exists

```
# remote feature exists; open PR titled "Fix login"
linked wt + origin/feature-pr + fake gh list → [{title:Fix login,url:…}]
  -> wrk --pr --title "Different title" --comment "please review"
  -> stderr: warning: title ignored (PR already exists); existing title: Fix login
  -> stdout: comment added\nURL
  -> no "pushed" line; no gh pr create; gh pr comment called
```

## Steps

1. Seed linked feature with remote head present.
2. Install fake gh; set list JSON to one open PR with title `Fix login`.
3. Run `--pr` with a **different** title and default comment.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	// Different title than existing — must be ignored.
	req.Args = prArgs("Different title", prDefaultComment)
	return nil
}
```

# Scenario

**Feature**: comment-only rejects empty/whitespace-only `--comment` after trim

```
linked wt -> wrk --pr --comment "   "   # no --title
  -> non-zero
  -> stderr mentions --comment (empty / required)
  -> no gh pr create / comment
```

## Steps

1. Seed linked feature + fake gh (open PR present so empty-check is not masked by “no open PR”).
2. Run `--pr --comment` with whitespace-only value and no `--title`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	req.Args = prCommentOnlyArgs("   ")
	return nil
}
```

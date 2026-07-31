# Scenario

**Feature**: `--pr --title` empty/whitespace-only is rejected (non-empty after trim)

```
linked wt -> wrk --pr --title "   " --comment "please review"
  -> non-zero
  -> stderr mentions --title (empty / required)
```

## Steps

1. Seed linked feature + fake gh.
2. Run `--pr` with whitespace-only title and valid comment.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	req.Args = prArgs("   ", prDefaultComment)
	return nil
}
```

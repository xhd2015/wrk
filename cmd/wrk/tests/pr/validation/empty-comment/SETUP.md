# Scenario

**Feature**: `--pr --comment` empty/whitespace-only is rejected (non-empty after trim)

```
linked wt -> wrk --pr --title "Fix login" --comment "   "
  -> non-zero
  -> stderr mentions --comment (empty / required)
```

## Steps

1. Seed linked feature + fake gh.
2. Run `--pr` with valid title and whitespace-only comment.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	req.Args = prArgs(prDefaultTitle, "   ")
	return nil
}
```

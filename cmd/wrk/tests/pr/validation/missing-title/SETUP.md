# Scenario

**Feature**: `--pr` without `--title` is rejected

```
linked wt -> wrk --pr --comment "please review"
  -> non-zero
  -> stderr mentions --title (required)
```

## Steps

1. Seed linked feature + fake gh (optional; validation may short-circuit).
2. Run `--pr --comment C` without `--title`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	req.Args = []string{"--pr", "--comment", prDefaultComment}
	return nil
}
```

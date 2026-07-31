# Scenario

**Feature**: incomplete create — `--pr --title T` without `--comment` is rejected

```
# title alone is not show mode and not a complete create
linked wt -> wrk --pr --title "Fix login"
  -> non-zero
  -> stderr mentions --comment (required for create)
```

## Steps

1. Seed linked feature + fake gh.
2. Run `--pr --title T` without `--comment` (incomplete create).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	req.Args = []string{"--pr", "--title", prDefaultTitle}
	return nil
}
```

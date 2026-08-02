# Scenario

**Feature**: Push DryRun with origin leaves remote branch tip unchanged

```
# main + bare origin (upstream set)
Caller -> Push({Checkout: main, DryRun: true})
  -> err nil; origin/main SHA unchanged
```

## Steps

1. Seed main with bare origin and upstream tracking.
2. Snapshot origin `refs/heads/main`.
3. Checkout = main; DryRun true from parent.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedPushRepoWithOrigin(t, req)
	req.Checkout = req.MainRepo
	req.OriginHEADBefore = revParseRef(t, req.OriginBare, "refs/heads/main")
	return nil
}
```

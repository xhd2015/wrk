# Scenario

**Feature**: TagNextAll DryRun plans multiple scope tags without creating refs

```
# root v0.0.1 + sub/v0.2.3; both scopes owned files changed at HEAD
Caller -> TagNextAll(main, DryRun)
  -> Tags includes v0.0.2 and sub/v0.2.4; no tag refs created
```

## Steps

1. Seed multi-scope bump repo (root + `sub/` tags; both paths changed).
2. Checkout = main; DryRun true from parent.
3. Run TagNextAll via OpTagNext (fills resp.Tags and resp.Tag).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMultiScopeBumpRepo(t, req)
	req.Checkout = req.MainRepo
	return nil
}
```

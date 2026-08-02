# Scenario

**Feature**: TagNext DryRun after root owned change plans v0.0.2

```
# v0.0.1 tagged; README changed at HEAD
Caller -> TagNext(main, DryRun) -> tag "v0.0.2"; refs/tags/v0.0.2 absent
```

## Steps

1. Seed root-bump repo (v0.0.1 + post-tag README change).
2. Checkout = main; DryRun true from parent.
3. Run TagNext.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedRootBumpRepo(t, req)
	req.Checkout = req.MainRepo
	return nil
}
```

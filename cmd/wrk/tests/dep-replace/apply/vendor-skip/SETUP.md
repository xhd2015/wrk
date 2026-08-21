# Scenario

**Feature**: apply writes absolute replace then skips tidy when vendor/ sits beside go.mod

```
nearest consumer has require + empty vendor/
  -> wrk --dep-replace <dep>
  -> replace example.com/dep => <abs>; skip tidy  (vendor/)
  -> no go.sum; vendor/ not rewritten
```

## Steps

1. Seed consumer with require + empty `vendor/` (nearest, not git).
2. Run apply.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVendorSkip(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```

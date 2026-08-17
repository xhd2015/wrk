# Scenario

**Feature**: dir-mode apply pins then skips tidy when vendor/ sits beside go.mod

```
nearest consumer has replace + require + empty vendor/
  -> wrk --dep-update <dep>
  -> dep-update example.com/dep -> v0.0.2
  -> skip tidy  module example.com/consumer  (vendor/)
  -> no go.sum; vendor/ not rewritten
```

## Steps

1. Seed drop-replace-latest + empty `vendor/` (nearest consumer, not git).
2. Run apply.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVendorSkipDir(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```

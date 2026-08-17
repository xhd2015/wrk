# Scenario

**Feature**: --all apply pins then skips tidy when vendor/ sits beside the consumer go.mod

```
# lib@v1.2.3 registered; app requires v1.0.0; app/vendor/ exists
cwd=app -> wrk --dep-update --all
  -> dep-update example.com/lib -> v1.2.3
  -> skip tidy  module example.com/app  (vendor/)
  -> summary updated 1; no go.sum; vendor/ not rewritten
```

## Steps

1. Seed cross-project outdated + modproxy + empty `vendor/`.
2. Run `--all` apply from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllVendorSkip(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```

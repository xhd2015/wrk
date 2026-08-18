# Scenario

**Feature**: dir-mode dry-run prints the stack tree and writes nothing

```
# primary + external/kool both require dep
cwd=primary -> wrk --dep-update <dep> --dry-run
  -> ==== dep-update (dry-run) ====
  -> would: pin on both checkouts; would: go mod tidy
  -> both go.mods unchanged; no go.sum
```

## Steps

1. Seed stack requirer other-checkout.
2. Run dry-run from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackOtherCheckoutRequirer(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```

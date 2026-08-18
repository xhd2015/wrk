# Scenario

**Feature**: dir-mode dry-run prints the stack tree and writes nothing

```
# primary + external/kool both require dep
cwd=primary -> wrk --dep-replace <dep> --dry-run
  -> ==== dep-replace (dry-run) ====
  -> would: replace on both checkouts
  -> both go.mods unchanged
```

## Steps

1. Seed stack other-checkout requirer.
2. Run dry-run from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackOtherCheckout(t, req)
	req.Args = []string{"--dep-replace", req.DepDir, "--dry-run"}
	return nil
}
```

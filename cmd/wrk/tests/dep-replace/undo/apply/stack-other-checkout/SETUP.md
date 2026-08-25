# Scenario

**Feature**: undo fans out across stack checkouts

```
primary + external/kool both require dep; HEAD no dep replace
WT: absolute dep replace on both
  -> wrk --dep-replace --undo
  -> drop on both checkouts
```

## Steps

1. Seed stack other-checkout (committed requires; no dep replace).
2. Introduce absolute dep replaces on primary and kool.
3. Run undo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackOtherCheckout(t, req)
	ensureVendorDir(t, req.ConsumerModDir)
	ensureVendorDir(t, req.Consumer2ModDir)
	appendAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	appendAbsoluteReplace(t, req.Consumer2GoMod, modDep, req.DepDir)
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.Args = []string{"--dep-replace", "--undo"}
	return nil
}
```

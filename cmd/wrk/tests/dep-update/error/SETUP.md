# Scenario

**Feature**: wrk --dep-update hard-fails on missing dir, non-module, no tags, no consumer, zero requirers

```
invalid dep | plain dir | untagged module | no go.mod ancestor | git module without require
  -> wrk --dep-update …
  -> Error non-zero
```

## Steps

- Leaves seed fixtures for each error class.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```

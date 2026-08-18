# Scenario

**Feature**: wrk root help documents --dep-update, partner --all, stack, and --dry-run

```
wrk -h
  -> root usage mentions --dep-update
  -> root usage mentions --all with --dep-update (inventory pull)
  -> root usage mentions unwind/stack + --dry-run for --dep-update
```

## Steps

- Descendants run help and assert flag documentation.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```

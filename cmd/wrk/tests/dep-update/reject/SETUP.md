# Scenario

**Feature**: wrk --dep-update rejects empty args and exclusive mode conflicts

```
wrk --dep-update (no dirs) | + --dep-replace | + --pin-locals
  -> non-zero
```

## Steps

- Descendants set conflicting or empty Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```

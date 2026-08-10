# Scenario

**Feature**: wrk --dep-replace rejects empty args and exclusive mode conflicts

```
wrk --dep-replace (no dirs) | + --dep-update | + --pin-locals
  -> non-zero
  -> requires directory / mutually exclusive
```

## Steps

- Descendants set conflicting or empty Args; no go.mod fixture required for mode conflicts.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReplaceHelpersUsed()
	return nil
}
```

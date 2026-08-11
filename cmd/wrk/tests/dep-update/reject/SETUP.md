# Scenario

**Feature**: wrk --dep-update rejects empty args, exclusive conflicts, and invalid --all forms

```
wrk --dep-update (no dirs, no --all)
  | + --dep-replace | + --pin-locals
  | --all without --dep-update
  | --dep-update --all + path args
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

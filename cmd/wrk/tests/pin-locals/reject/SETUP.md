# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with other modes and partners

```
wrk --pin-locals + (--done|--unwind|--bring|--list|--commit|--add-all)
  -> non-zero
  -> mutually exclusive (or equivalent)
```

## Steps

- Descendants combine `--pin-locals` with a forbidden flag.
- Exclusion should fire before planning (no git/go.mod required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensurePinLocalsHelpersUsed()
	return nil
}
```

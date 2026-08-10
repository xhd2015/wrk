# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with --unwind

```
wrk --pin-locals --unwind
  -> non-zero
  -> mutually exclusive (or equivalent exclusive wording)
```

## Steps

1. Run `wrk --pin-locals --unwind` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--unwind"}
	return nil
}
```

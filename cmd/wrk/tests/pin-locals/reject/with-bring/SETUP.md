# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with --bring

```
wrk --pin-locals --bring
  -> non-zero
  -> mutually exclusive (or equivalent exclusive wording)
```

## Steps

1. Run `wrk --pin-locals --bring` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--bring", req.WorkRoot}
	return nil
}
```

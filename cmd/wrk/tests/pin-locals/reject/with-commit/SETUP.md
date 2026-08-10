# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with --commit

```
wrk --pin-locals --commit
  -> non-zero
  -> mutually exclusive (or equivalent exclusive wording)
```

## Steps

1. Run `wrk --pin-locals --commit` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--commit"}
	return nil
}
```

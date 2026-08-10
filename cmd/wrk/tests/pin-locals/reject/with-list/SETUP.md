# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with --list

```
wrk --pin-locals --list
  -> non-zero
  -> mutually exclusive (or equivalent exclusive wording)
```

## Steps

1. Run `wrk --pin-locals --list` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--list"}
	return nil
}
```

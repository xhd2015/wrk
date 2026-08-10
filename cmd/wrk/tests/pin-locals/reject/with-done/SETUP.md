# Scenario

**Feature**: wrk --pin-locals is mutually exclusive with --done

```
wrk --pin-locals --done
  -> non-zero
  -> mutually exclusive (or equivalent exclusive wording)
```

## Steps

1. Run `wrk --pin-locals --done` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--done"}
	return nil
}
```

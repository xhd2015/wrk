# Scenario

**Feature**: wrk --dep-replace is mutually exclusive with --pin-locals

```
wrk --dep-replace <dir> --pin-locals
  -> non-zero
  -> mutually exclusive
```

## Steps

1. Combine `--dep-replace` with `--pin-locals` (exclusive family).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-replace", req.WorkRoot, "--pin-locals"}
	return nil
}
```

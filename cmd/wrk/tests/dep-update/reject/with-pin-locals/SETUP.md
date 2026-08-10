# Scenario

**Feature**: wrk --dep-update is mutually exclusive with --pin-locals

```
wrk --dep-update <dir> --pin-locals
  -> non-zero
  -> mutually exclusive
```

## Steps

1. Combine `--dep-update` with `--pin-locals`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-update", req.WorkRoot, "--pin-locals"}
	return nil
}
```

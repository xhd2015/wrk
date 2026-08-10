# Scenario

**Feature**: wrk --dep-update is XOR with --dep-replace

```
wrk --dep-update <dir> --dep-replace <dir>
  -> non-zero
  -> mutually exclusive
```

## Steps

1. Combine both flags with placeholder paths.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-update", req.WorkRoot, "--dep-replace", req.WorkRoot}
	return nil
}
```

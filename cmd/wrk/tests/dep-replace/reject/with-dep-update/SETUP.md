# Scenario

**Feature**: wrk --dep-replace is XOR with --dep-update

```
wrk --dep-replace <dir> --dep-update <dir>
  -> non-zero
  -> mutually exclusive
```

## Steps

1. Combine both flags with placeholder paths (mode conflict before path validation OK).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Paths need not exist: exclusivity should fire first.
	req.Args = []string{"--dep-replace", req.WorkRoot, "--dep-update", req.WorkRoot}
	return nil
}
```

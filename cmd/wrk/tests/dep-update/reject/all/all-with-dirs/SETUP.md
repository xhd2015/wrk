# Scenario

**Feature**: --dep-update --all with path args is rejected

```
wrk --dep-update --all <path>
  -> non-zero
  -> cannot combine --all with directory args
```

## Steps

1. Run `wrk --dep-update --all` with a path remaining arg.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Placeholder path; must not be treated as dir-mode update.
	req.Args = []string{"--dep-update", "--all", req.WorkRoot}
	return nil
}
```

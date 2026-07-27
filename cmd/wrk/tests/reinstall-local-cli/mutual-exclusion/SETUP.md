# Scenario

**Feature**: wrk --reinstall-local is mutually exclusive with other modes

```
wrk --reinstall-local + another mode flag -> non-zero, mutually exclusive
```

## Steps

- Descendants combine `--reinstall-local` with another wrk mode flag.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: exclusive-mode error path.
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```

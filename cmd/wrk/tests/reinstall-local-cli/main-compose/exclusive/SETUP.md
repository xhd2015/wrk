# Scenario

**Feature**: --main --reinstall-local compose remains exclusive with other modes

```
# compose does not open the door to stacking unrelated modes
wrk --main --reinstall-local --list
  -> non-zero, mutually exclusive
```

## Steps

- Descendants combine compose flags with another mode (`--list`).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: exclusive error path after compose is allowed.
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```

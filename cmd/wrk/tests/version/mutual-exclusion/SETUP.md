# Scenario

**Feature**: wrk --version is mutually exclusive with other wrk modes

```
wrk --version + another mode flag -> non-zero, mutually exclusive
```

## Steps

- Descendants combine `--version` with another wrk mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	ensureVersionHelpersUsed()
	return nil
}
```
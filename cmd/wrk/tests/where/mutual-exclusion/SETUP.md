# Scenario

**Feature**: wrk --where is mutually exclusive with other modes

```
wrk --where spl + another mode flag -> non-zero, mutually exclusive
```

## Steps

- Descendants combine `--where` with another wrk mode flag.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
	return nil
}```

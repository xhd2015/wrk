# Scenario

**Feature**: wrk --where requires a basename argument

```
wrk --where (no value) -> non-zero exit, requires argument error
```

## Steps

- Descendants invoke `wrk --where` without a following basename argument.

```go
func Setup(t *testing.T, req *Request) error {
	ensureWhereHelpersUsed()
	return nil
}```

# Scenario

**Feature**: wrk --dep-update with zero directory args errors

```
wrk --dep-update
  -> non-zero
  -> requires directory
```

## Steps

1. Run `wrk --dep-update` with no path args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-update"}
	return nil
}
```

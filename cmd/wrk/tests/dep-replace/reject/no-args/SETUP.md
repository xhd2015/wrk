# Scenario

**Feature**: wrk --dep-replace with zero directory args and no --undo errors

```
wrk --dep-replace
  -> non-zero
  -> Error requires a directory or --undo
```

## Steps

1. Run `wrk --dep-replace` with no path args from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--dep-replace"}
	return nil
}
```

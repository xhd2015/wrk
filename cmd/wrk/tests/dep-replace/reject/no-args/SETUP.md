# Scenario

**Feature**: wrk --dep-replace with zero directory args errors

```
wrk --dep-replace
  -> non-zero
  -> Error requires directory (or equivalent empty-path wording)
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

# Scenario

**Feature**: wrk -h documents --dep-update

```
wrk -h
  -> exit 0
  -> help text contains --dep-update
```

## Steps

1. Run `wrk -h` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

# Scenario

**Feature**: wrk -h documents --dep-replace

```
wrk -h
  -> exit 0
  -> help text contains --dep-replace
```

## Steps

1. Run `wrk -h` from neutral WorkRoot (cwd irrelevant for help).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

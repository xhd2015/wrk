# Scenario

**Feature**: wrk -h documents --all as partner of --dep-update

```
wrk -h
  -> exit 0 (or help success)
  -> help text contains --all in context of --dep-update
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

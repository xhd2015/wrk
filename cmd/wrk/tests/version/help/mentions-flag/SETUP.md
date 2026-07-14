# Scenario

**Feature**: wrk -h documents --version

```
wrk -h
  -> exit 0
  -> help text contains --version
```

## Steps

1. Run `wrk -h` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
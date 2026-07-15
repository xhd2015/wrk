# Scenario

**Feature**: wrk -h documents --scan-git-repos and --no-cache

```
wrk -h
  -> exit 0
  -> help text contains --scan-git-repos and --no-cache
```

## Steps

1. Run `wrk -h` from isolated WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```

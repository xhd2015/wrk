# Scenario

**Feature**: wrk -h documents --reinstall-local

```
# C7
wrk -h
  -> exit 0
  -> help text contains --reinstall-local
```

## Steps

1. Run `wrk -h` from neutral module dir.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```

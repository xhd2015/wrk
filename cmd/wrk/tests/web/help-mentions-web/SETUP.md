# Scenario

**Feature**: wrk -h documents --web and --port

```
wrk -h
  -> exit 0
  -> help text contains --web (and ideally --port)
```

## Steps

1. Run `wrk -h` from isolated WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```

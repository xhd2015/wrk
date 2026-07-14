# Scenario

**Feature**: wrk --rm without path argument errors

```
wrk --rm (no path) -> non-zero exit
```

## Steps

1. Run `wrk --rm` with no following path argument.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--rm"}
	return nil
}
```

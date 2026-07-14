# Scenario

**Feature**: wrk --cd without path argument errors

```
wrk --cd (no path) -> non-zero; wrk: --cd requires a path argument
```

## Steps

1. Run `wrk --cd` only.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--cd"}
	req.TargetDir = ""
	return nil
}
```

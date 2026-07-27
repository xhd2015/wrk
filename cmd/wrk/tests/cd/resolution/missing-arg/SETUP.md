# Scenario

**Feature**: wrk --cd without path argument errors

```
wrk --cd (no path) -> non-zero; wrk: --cd requires a path argument
```

## Steps

1. Run `wrk --cd` only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--cd"}
	req.TargetDir = ""
	return nil
}
```

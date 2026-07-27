# Scenario

**Feature**: wrk --add without path argument errors

```
wrk --add (no path) -> non-zero exit
```

## Steps

1. Run `wrk --add` with no following path argument.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--add"}
	return nil
}
```
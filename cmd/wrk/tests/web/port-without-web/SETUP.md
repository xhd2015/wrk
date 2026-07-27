# Scenario

**Feature**: wrk --port without --web is rejected

```
wrk --port 18080 (no --web)
  -> non-zero exit
  -> stderr: --port is only valid with --web
  -> stdout empty
```

## Steps

1. Run `wrk --port 18080` from isolated WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Fixed high port is fine: process must reject before bind.
	req.Args = []string{"--port", "18080"}
	return nil
}
```

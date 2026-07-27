# Scenario

**Feature**: wrk --web --status is mutually exclusive

```
wrk --web --status -> non-zero; mutually exclusive; empty stdout
```

## Steps

1. Run `wrk --web --status` from isolated WorkRoot (no git required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--web", "--status"}
	return nil
}
```

# Scenario

**Feature**: wrk --web --list is mutually exclusive

```
wrk --web --list -> non-zero; mutually exclusive; empty stdout
```

## Steps

1. Run `wrk --web --list` from isolated WorkRoot (no git required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--web", "--list"}
	return nil
}
```

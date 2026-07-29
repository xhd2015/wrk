# Scenario

**Feature**: bare wrk --no-dep is rejected

```
wrk --no-dep -> non-zero
  stderr contains: --no-dep is only valid with --dep, --bring, or --all-deps
```

## Steps

1. Run `wrk --no-dep` from neutral WorkRoot (no git fixtures required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"--no-dep"}
	return nil
}
```

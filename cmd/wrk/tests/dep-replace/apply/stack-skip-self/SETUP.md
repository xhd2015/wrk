# Scenario

**Feature**: dep’s own go.mod is not rewritten when the dep checkout is on the stack

```
# primary requires dep; replace dep => ./external/dep (own git)
cwd=primary -> wrk --dep-replace <external/dep>
  -> replace on example.com/app only
  -> example.com/dep go.mod unchanged (self)
```

## Steps

1. Seed primary replace to on-stack dep checkout.
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackSkipSelf(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```

# Scenario

**Feature**: nested monorepo dep uses tag prefix packages/dep/vN.N.N → version vN.N.N

```
monorepo packages/dep (no root go.mod); tags packages/dep/v0.0.1, packages/dep/v0.0.2
consumer has replace + require v0.0.1
  -> wrk --dep-update <packages/dep>
  -> dep-update example.com/dep -> v0.0.2
  -> go mod tidy ok  module example.com/consumer
  -> optional tag form may mention packages/dep/v0.0.2
  -> replace dropped; require v0.0.2; go.sum exists
```

## Steps

1. Seed nested tag-prefix fixture + file:// GOPROXY.
2. Run apply.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupNestedTagPrefix(t, req)
	enableDirModeTidyProxy(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```

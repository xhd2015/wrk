# Scenario

**Feature**: leftover `create.interceptor` is ignored; native create runs

```
config has only create.interceptor enabled
wrk -> native create; no intercept; exit 0
```

## Steps

- Seed interceptor-only config; run create.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```

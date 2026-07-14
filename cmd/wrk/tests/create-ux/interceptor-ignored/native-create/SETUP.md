# Scenario

**Feature**: interceptor-only config does not intercept create

```
create.interceptor enabled leftover -> wrk -> native wt path; UX mocks silent
```

## Steps

1. Write interceptor-only config.
2. Run bare wrk.

```go
func Setup(t *testing.T, req *Request) error {
	writeInterceptorOnlyConfig(t, req.WrkHome)
	req.Args = nil
	return nil
}
```

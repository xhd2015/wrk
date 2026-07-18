# Scenario

**Feature**: interceptor-only config does not intercept create via `--new`

```
create.interceptor enabled leftover -> wrk --new -> native wt path; UX mocks silent
```

## Steps

1. Write interceptor-only config.
2. Run `wrk --new`.

```go
func Setup(t *testing.T, req *Request) error {
	writeInterceptorOnlyConfig(t, req.WrkHome)
	req.Args = []string{"--new"}
	return nil
}
```

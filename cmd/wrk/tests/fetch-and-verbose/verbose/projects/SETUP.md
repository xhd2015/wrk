# Scenario

**Feature**: wrk --projects -v logs fetch only when --fetch is set

```
--projects -v (no fetch) -> stderr empty (all minor reads)
--projects --fetch -v -> stderr contains fetch, no rev-parse/status
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
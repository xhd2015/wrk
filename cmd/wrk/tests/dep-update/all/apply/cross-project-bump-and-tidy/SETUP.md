# Scenario

**Feature**: --all apply bumps outdated inventory require and runs go mod tidy

```
# lib@v1.2.3 registered; app requires v1.0.0; file:// proxy for tidy
cwd=app -> wrk --dep-update --all
  -> ==== dep-update ====; pin example.com/lib; no argv dep list
  -> go mod tidy ok
  -> dep-update: updated 1, already 0, skipped 0 in 1 checkouts
  -> app require@v1.2.3; go.sum exists; owner go.mod unchanged
```

## Steps

1. Seed cross-project outdated + modproxy.
2. Run apply from app (consumer not registered).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllWithProxy(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```

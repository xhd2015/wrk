# Scenario

**Feature**: install dry-run prints would-* report without writing

```
wrk --bash-integration --install --dry-run
  -> stdout: bash integration: would install|would update|is up to date
  -> no filesystem writes
```

## Steps

1. Set `req.Mode = "install"` and `req.DryRun = true`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "install"
	req.DryRun = true
	return nil
}
```

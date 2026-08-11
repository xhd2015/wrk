# Scenario

**Feature**: --all silently skips non-inventory requires; bumps inventory-owned

```
# app requires example.com/external@v9.9.9 (unknown) + lib@v1.0.0 (owned)
cwd=app -> wrk --dep-update --all
  -> dep-update example.com/lib -> v1.2.3 only
  -> no stdout line for external
  -> skipped 0 (external silent, not counted)
```

## Steps

1. Seed app with external + outdated lib; register only lib; modproxy.
2. Run apply from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllSkipExternal(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```

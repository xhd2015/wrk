# Scenario

**Feature**: --all resolves nested owner module tags (packages/dep/vN → vN)

```
# owner monorepo packages/dep module example.com/lib/dep; tags packages/dep/v0.1.0,v0.2.0
# app requires example.com/lib/dep@v0.1.0
cwd=app -> wrk --dep-update --all
  -> dep-update example.com/lib/dep -> v0.2.0
  -> optional tag packages/dep/v0.2.0
  -> tidy ok; require@v0.2.0
```

## Steps

1. Seed nested owner module + outdated app; register owner monorepo; modproxy.
2. Run apply from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllNestedOwnerModuleTag(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```

# Scenario

**Feature**: --all --dry-run plans bump of outdated inventory-owned require + tidy

```
# lib tagged v1.0.0,v1.2.3 registered; app requires example.com/lib@v1.0.0
cwd=app -> wrk --dep-update --all --dry-run
  -> would: dep-update example.com/lib -> v1.2.3
  -> would: go mod tidy  module example.com/app
  -> dep-update: would update 1, already 0, skipped 0
  -> app go.mod unchanged; owner go.mod unchanged
```

## Steps

1. Seed owner lib + outdated app consumer; register owner only.
2. Run dry-run from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllCrossProjectOutdated(t, req)
	req.Args = []string{"--dep-update", "--all", "--dry-run"}
	return nil
}
```

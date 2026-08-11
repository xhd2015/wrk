# Scenario

**Feature**: --all --dry-run when all inventory-owned requires already at latest

```
# lib tagged v1.2.3; app already requires example.com/lib@v1.2.3
cwd=app -> wrk --dep-update --all --dry-run
  -> dep-update: already up to date
  -> dep-update: updated 0, already 1, skipped 0
  -> no would: dep-update pin lines
  -> go.mod unchanged
```

## Steps

1. Seed owner + already-current app; register owner.
2. Run dry-run from app.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllAlreadyCurrent(t, req)
	req.Args = []string{"--dep-update", "--all", "--dry-run"}
	return nil
}
```

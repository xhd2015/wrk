# Scenario

**Feature**: wrk --dep-update --dry-run plans pin without writing

```
consumer + tagged dep with replace
  -> wrk --dep-update <dep> --dry-run
  -> would: dep-update …; go.mod unchanged
```

## Steps

- Leaves seed fixtures and dry-run Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```

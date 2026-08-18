# Scenario

**Feature**: wrk --dep-replace applies absolute replaces (nearest or stack)

```
consumer + dep module(s)
  -> wrk --dep-replace <dir>…
  -> ==== dep-replace ====; checkout → module → replace
  -> go.mod gains absolute replace; no tidy
```

## Steps

- Leaves seed fixtures and set apply Args (no --dry-run).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReplaceHelpersUsed()
	return nil
}
```

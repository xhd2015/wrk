# Scenario

**Feature**: wrk --dep-replace --dry-run plans absolute replaces without writing

```
consumer + dep
  -> wrk --dep-replace <dep> --dry-run
  -> would: dep-replace …; go.mod unchanged; no tidy
```

## Steps

- Leaves seed consumer/dep and set dry-run Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReplaceHelpersUsed()
	return nil
}
```

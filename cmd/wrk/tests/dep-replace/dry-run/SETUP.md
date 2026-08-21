# Scenario

**Feature**: wrk --dep-replace --dry-run plans absolute replaces without writing

```
consumer + dep
  -> wrk --dep-replace <dep> --dry-run
  -> ==== dep-replace (dry-run) ====; would: replace + would: go mod tidy; go.mod unchanged
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

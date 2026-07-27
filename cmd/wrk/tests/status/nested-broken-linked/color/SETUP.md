# Scenario

**Feature**: --color highlights red error value on scan-discovered broken block

```
# same nested broken linked wt fixture as stale-gitdir
vendor/host/broken-wt (broken) + wrk --status --color -> red error: value on scan block
```

## Steps

1. Build fixture via `setupNestedBrokenLinkedFixture`.
2. Run `wrk --status --color` from `{WorkRoot}/myrepo` (pipe-safe).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureNestedBrokenLinkedHelpersUsed()
	setupNestedBrokenLinkedFixture(t, req)
	req.Args = []string{"--status", "--color"}
	return nil
}
```
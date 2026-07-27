# Scenario

**Bug**: scan-discovered linked worktree with stale gitdir aborts wrk --status

```
# nested linked wt under vendor/host has broken git metadata
vendor/host/broken-wt (stale gitdir) -> scan discovers row -> enrich fails

# today: printStatusBlock returns error and aborts entire run
enrich error -> fatal exit (healthy sibling blocks missing)

# desired: minimal broken block, continue, exit 0
broken scan block -> Dir: vendor/host/broken-wt + Status: error: ... ; siblings still print
```

## Steps

1. Build fixture via `setupNestedBrokenLinkedFixture` (main + tools/good + vendor/host + broken-wt).
2. Run `wrk --status` from `{WorkRoot}/myrepo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureNestedBrokenLinkedHelpersUsed()
	setupNestedBrokenLinkedFixture(t, req)
	return nil
}
```
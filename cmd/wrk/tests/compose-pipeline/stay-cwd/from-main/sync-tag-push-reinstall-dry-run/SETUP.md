# Scenario

**Feature**: on main without done — `--sync --tag-next --push --reinstall-local` composes; fixed order; all on main activeRoot

```
# Main has owned change + origin + present stub
myrepo (main)
  -> wrk --sync --tag-next --push --reinstall-local --dry-run
  -> NOT mutually exclusive
  -> plans: sync → tag-next → push → reinstall (activeRoot stays main)
  -> exit 0; zero mutations
```

## Steps

1. Main+origin with owned change; seed reinstall; baseline.
2. Run multi-stage dry-run (no done).

```go
func Setup(t *testing.T, req *Request) error {
	setupAPMainOnOrigin(t, req)
	_ = seedAPReinstallPresent(t, req)
	// Re-push tip after present commit so origin baseline matches (optional for dry-run).
	recordAPDryRunBaseline(t, req)
	req.Args = []string{
		"--sync", "--tag-next", "--push", "--reinstall-local", "--dry-run",
	}
	return nil
}
```

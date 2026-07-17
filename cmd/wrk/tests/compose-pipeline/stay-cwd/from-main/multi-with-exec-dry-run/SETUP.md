# Scenario

**Feature**: `--exec` is valid as last stage of multi-stage compose without `--done` (final activeRoot = main)

```
myrepo (main)
  -> wrk --sync --tag-next --dry-run --exec true
  -> NOT mutually exclusive
  -> NOT "--exec is only valid with --done"
  -> tag/sync plans then exec as last stage
```

## Steps

1. Main+origin owned change; baseline.
2. Run multi-stage + exec dry-run.

```go
func Setup(t *testing.T, req *Request) error {
	setupAPMainOnOrigin(t, req)
	recordAPDryRunBaseline(t, req)
	req.Args = []string{"--sync", "--tag-next", "--dry-run", "--exec", "true"}
	return nil
}
```

# Scenario

**Feature**: JSON dry-run output plans v0.0.2 without creating tags

```
# --tag-next --dry-run --json -> JSON on stdout, no ANSI, no git tag
git repo + tags -> wrk --tag-next --dry-run --json -> machine-readable plan
```

## Steps

1. `setupRootBumpRepo`.
2. Run `wrk --tag-next --dry-run --json`.

```go
func Setup(t *testing.T, req *Request) error {
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next", "--dry-run", "--json"}
	return nil
}
```
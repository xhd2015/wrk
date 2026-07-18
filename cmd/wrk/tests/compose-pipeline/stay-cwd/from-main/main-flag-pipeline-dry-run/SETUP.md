# Scenario

**Feature**: already on main + `--main` + pipeline dry-run — notice that `--main` is not necessary; continue pipeline

```
# cwd is main checkout with owned change + origin + present stub
myrepo (main)
  -> wrk --main --sync --tag-next --push --reinstall-local --dry-run
  -> stderr notice: --main is not necessary (already at main …); continuing
  -> NOT mutually exclusive
  -> plans still run on main; exit 0; zero mutations
```

## Steps

1. Main+origin with owned change; seed reinstall; baseline.
2. Run multi-stage dry-run with redundant `--main`.

```go
func Setup(t *testing.T, req *Request) error {
	setupAPMainOnOrigin(t, req)
	_ = seedAPReinstallPresent(t, req)
	recordAPDryRunBaseline(t, req)
	req.Args = []string{
		"--main",
		"--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
	}
	return nil
}
```

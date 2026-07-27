# Scenario

**Feature**: full ship with `--exec` as last stage after done path; exec plans/runs in final activeRoot (main)

```
# Linked wt + full posts + --exec true --dry-run
linked wt
  -> wrk --done --sync --tag-next --push --reinstall-local --dry-run --exec true
  -> done plan then posts on would-be main
  -> --exec is last pipeline stage (accepted; not rejected as mode conflict)
  -> dry-run: either would-exec vocabulary or at least no mutex / no "exec only with"
  -> zero mutations
```

## Steps

1. Linked ahead + origin; reinstall seed; baseline.
2. Run done multi-stage with `--exec true` and `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAPLinkedAheadOrigin(t, req)
	_ = seedAPReinstallPresent(t, req)
	recordAPDryRunBaseline(t, req)
	req.Args = []string{
		"--done", "--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
		"--exec", "true",
	}
	return nil
}
```

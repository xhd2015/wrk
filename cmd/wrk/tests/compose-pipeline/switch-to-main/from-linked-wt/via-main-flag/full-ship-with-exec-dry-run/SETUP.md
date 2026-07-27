# Scenario

**Feature**: linked worktree `--main` pipeline includes `--exec` as last stage on main (dry-run)

```
# Linked wt + origin + present stub
linked wt
  -> wrk --main --sync --tag-next --push --reinstall-local --dry-run --exec true
  -> NOT mutually exclusive
  -> --exec accepted under main activeRoot (not mode conflict)
  -> dry-run: plans/posts on main; zero mutations; wt kept
```

## Steps

1. Linked ahead + origin; reinstall seed; baseline.
2. Run `--main` multi-stage with `--exec true` and `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAPLinkedAheadOrigin(t, req)
	_ = seedAPReinstallPresent(t, req)
	recordAPDryRunBaseline(t, req)
	req.Args = []string{
		"--main",
		"--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
		"--exec", "true",
	}
	return nil
}
```

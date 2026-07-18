# Scenario

**Feature**: linked worktree full ship via `--main` (no `--done`): sync → tag-next → push → reinstall on main dry-run

```
# Linked wtA ahead; wtB; origin; GOBIN present stub
linked wt
  -> wrk --main --sync --tag-next --push --reinstall-local --dry-run
  -> NOT mutually exclusive
  -> activeRoot := main (scope rewrite; no nested shell)
  -> plans: sync → tag-next → push → reinstall on main
  -> no done/merge-back plan lines; worktree kept
  -> exit 0; zero mutations
```

## Steps

1. Sync+origin fixture (wtA ahead, wtB stays); seed reinstall on main; dry-run baseline.
2. Run `--main` multi-stage dry-run from linked wt cwd (no `-y`).

```go
func Setup(t *testing.T, req *Request) error {
	setupAPSyncWithOrigin(t, req)
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

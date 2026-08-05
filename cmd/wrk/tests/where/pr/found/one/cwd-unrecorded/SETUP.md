# Scenario

**Feature**: main not pre-recorded in projects.json; cwd is linked wt with matching origin → still finds path

```
# no wrk --add; cwd = linked wt of acme/app on feature-pr
cwd linked wt (unrecorded main)
  -> wrk --where --pr https://github.com/acme/app/pull/42
  -> cwd main origin matches -> worktree on head
  -> stdout = linked abs path
```

## Steps

1. Seed main + linked on head; do **not** pre-record via `--add`.
2. Install fake gh; run from linked worktree as cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupUnrecordedLinked(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```

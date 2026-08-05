# Scenario

**Feature**: PR URL resolves via projects.json main (neutral cwd) to one linked worktree

```
# main recorded; origin matches acme/app; linked wt on feature-pr
projects.json + neutral cwd
  -> wrk --where --pr https://github.com/acme/app/pull/42
  -> gh pr view 42 --repo acme/app --json headRefName
  -> stdout = linked abs path
```

## Steps

1. Seed main with github origin + linked worktree on head; record via `wrk --add`.
2. Install fake gh with headRefName=feature-pr.
3. Run `--where --pr <full-url>` from neutral non-git cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```

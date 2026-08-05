# Scenario

**Feature**: matching main exists but head branch not checked out anywhere → error

```
recorded main (origin acme/app) on main only; no feature-pr worktree
  -> wrk --where --pr URL
  -> non-zero; empty stdout
  -> stderr names PR number, head branch (feature-pr), and repo
```

## Steps

1. Seed recorded main with matching origin; do not create head worktree.
2. Fake gh returns headRefName=feature-pr.
3. Run from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupRecordedMainOnly(t, req)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```

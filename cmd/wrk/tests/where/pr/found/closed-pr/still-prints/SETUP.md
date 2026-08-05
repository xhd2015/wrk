# Scenario

**Feature**: closed PR still prints local worktree path for head branch

```
recorded linked wt on feature-pr; FAKE_GH_VIEW_JSON state=CLOSED
  -> wrk --where --pr URL
  -> stdout linked path (not refused as closed)
```

## Steps

1. Seed recorded linked fixture + fake gh.
2. Override view JSON to CLOSED with headRefName.
3. Run default `--where --pr` URL.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	setWherePrViewJSON(t, req, wherePrViewJSON(wherePrHeadBranch, "CLOSED"))
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```

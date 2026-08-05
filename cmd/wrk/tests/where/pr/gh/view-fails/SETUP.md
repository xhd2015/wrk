# Scenario

**Feature**: `gh pr view` failure surfaces as non-zero error

```
recorded linked + fake gh view exits 1
  -> wrk --where --pr URL
  -> non-zero; empty stdout
  -> stderr surfaces gh / view failure
```

## Steps

1. Seed recorded linked + fake gh.
2. Set FAKE_GH_VIEW_EXIT=1.
3. Run default compose argv.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	setWherePrViewExit(t, req, 1)
	req.Args = wherePrArgs(wherePrURL)
	return nil
}
```

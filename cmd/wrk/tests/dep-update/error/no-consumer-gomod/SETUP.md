# Scenario

**Feature**: --dep-update fails when no go.mod is found walking up from workDir

```
tagged dep present; workDir has no go.mod ancestor
  -> wrk --dep-update <dep>
  -> non-zero
```

## Steps

1. Seed dep-only tagged fixture (RepoDir = WorkRoot).
2. Run update.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDepOnlyTagged(t, req)
	req.Args = []string{"--dep-update", req.DepDir}
	return nil
}
```

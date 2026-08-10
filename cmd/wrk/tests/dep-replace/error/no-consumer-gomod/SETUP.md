# Scenario

**Feature**: --dep-replace fails when no go.mod is found walking up from workDir

```
dep module present; workDir has no go.mod ancestor under WorkRoot
  -> wrk --dep-replace <dep>
  -> non-zero
```

## Steps

1. Seed dep-only fixture (RepoDir = WorkRoot without go.mod).
2. Run replace with valid dep path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDepOnly(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```

# Scenario

**Feature**: optional dir filter drops only matching introduced OldPaths

```
WT: replace dep + dep2 (both introduced)
  -> wrk --dep-replace --undo <dep>
  -> drop dep only; dep2 replace remains
```

## Steps

1. Seed two introduced replaces.
2. Undo with only dep dir filter.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUndoTwoIntroduced(t, req)
	req.Args = []string{"--dep-replace", "--undo", req.DepDir}
	return nil
}
```

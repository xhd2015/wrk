# Scenario

**Feature**: undo keeps replaces already on HEAD; drops only introduced

```
HEAD: replace kool => ./external/kool
WT: + replace dep => abs
  -> wrk --dep-replace --undo
  -> drop dep only; kool replace remains
```

## Steps

1. Seed git primary with committed kool replace.
2. Introduce absolute dep replace in WT.
3. Run undo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUndoKeepsHead(t, req)
	req.Args = []string{"--dep-replace", "--undo"}
	return nil
}
```

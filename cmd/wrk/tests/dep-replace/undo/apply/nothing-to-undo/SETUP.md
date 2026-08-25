# Scenario

**Feature**: undo is soft no-op when WT matches HEAD replaces

```
git HEAD == WT (no introduced replace)
  -> wrk --dep-replace --undo
  -> dep-replace: nothing to undo; exit 0; go.mod unchanged
```

## Steps

1. Seed clean git primary matching HEAD.
2. Run undo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUndoNothing(t, req)
	req.Args = []string{"--dep-replace", "--undo"}
	return nil
}
```

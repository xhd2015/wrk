# Scenario

**Feature**: undo drops a replace introduced since HEAD

```
git HEAD: require dep, no replace
WT: replace dep => abs
  -> wrk --dep-replace --undo
  -> drop dep replace; require version unchanged; tidy ok
```

## Steps

1. Seed git primary with committed require (no replace).
2. Add absolute replace in working tree.
3. Run `--dep-replace --undo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUndoIntroduced(t, req)
	req.Args = []string{"--dep-replace", "--undo"}
	return nil
}
```

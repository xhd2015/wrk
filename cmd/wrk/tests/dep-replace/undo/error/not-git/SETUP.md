# Scenario

**Feature**: undo requires git HEAD

```
not-git consumer
  -> wrk --dep-replace --undo
  -> non-zero; wrk: … requires git HEAD; no banner
```

## Steps

1. Seed nearest not-git consumer+dep.
2. Run undo.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.Args = []string{"--dep-replace", "--undo"}
	return nil
}
```

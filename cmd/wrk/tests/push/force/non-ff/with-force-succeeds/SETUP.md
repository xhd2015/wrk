# Scenario

**Feature**: `wrk --push -f` force-with-lease overwrites diverged origin tip

```
diverged main vs origin/main
  -> wrk --push -f
  -> exit 0
  -> pushed main → origin/main
  -> origin/main == local HEAD (local-only tip wins)
```

## Steps

1. Build diverged main + origin fixture.
2. Run `wrk --push -f`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushDivergedMainWithOrigin(t, req)
	req.Args = []string{"--push", "-f"}
	return nil
}
```

# Scenario

**Feature**: bare `wrk --push` fails on diverged origin (control for force path)

```
diverged main vs origin/main
  -> wrk --push
  -> non-zero
  -> origin/main stays at pre-run remote-only tip
  -> no success confirm "pushed main → origin/main"
```

## Steps

1. Build diverged main + origin fixture.
2. Run `wrk --push` (no force).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushDivergedMainWithOrigin(t, req)
	req.Args = []string{"--push"}
	return nil
}
```

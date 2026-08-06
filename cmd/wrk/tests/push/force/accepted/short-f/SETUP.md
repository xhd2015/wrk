# Scenario

**Feature**: `wrk --push -f` accepts short force and pushes with force-with-lease semantics

```
myrepo (main) + origin
  -> wrk --push -f
  -> pushed main → origin/main
  -> origin/main == local HEAD
```

## Steps

1. Seed main with bare origin (upstream set; tips equal after setup).
2. Run `wrk --push -f`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push", "-f"}
	return nil
}
```

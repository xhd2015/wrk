# Scenario

**Feature**: bare `wrk --push` without origin/upstream fails clearly

```
# main checkout, no remotes
myrepo (main, no origin)
  -> wrk --push
  -> non-zero
  -> stderr mentions no upstream / no origin
  -> no "pushed …" success line
```

## Steps

1. Seed main repo without remotes.
2. Run `wrk --push`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushNoRemote(t, req)
	req.Args = []string{"--push"}
	return nil
}
```

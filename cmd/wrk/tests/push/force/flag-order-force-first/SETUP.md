# Scenario

**Feature**: flag order free — `-f --push` apply same as `--push -f`

```
myrepo (main) + origin
  -> wrk -f --push
  -> pushed main → origin/main
  -> origin/main == local HEAD
```

## Steps

1. Seed main with bare origin.
2. Run `wrk -f --push` (force before push).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"-f", "--push"}
	return nil
}
```

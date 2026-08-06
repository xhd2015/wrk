# Scenario

**Feature**: `wrk --push --force` accepts long force form

```
myrepo (main) + origin
  -> wrk --push --force
  -> pushed main → origin/main
  -> origin/main == local HEAD
```

## Steps

1. Seed main with bare origin.
2. Run `wrk --push --force`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push", "--force"}
	return nil
}
```

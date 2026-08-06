# Scenario

**Feature**: `--push` + force accepted when remote tip is fast-forward-ok (or already equal)

```
myrepo (main) + origin (FF-ok / equal tips)
  -> wrk --push -f | --push --force
  -> exit 0
  -> pushed main → origin/main
  -> origin/main == local HEAD
```

## Steps

- Grouping: leaves use `setupPushMainWithOrigin` and set force+push Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.InProcess = true
	return nil
}
```

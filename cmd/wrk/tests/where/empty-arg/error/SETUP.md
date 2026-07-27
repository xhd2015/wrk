# Scenario

**Feature**: wrk --where without argument errors

```
wrk --where (no value) -> wrk: --where requires a path argument
```

## Steps

1. Run `wrk --where` with no following basename argument from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--where"}
	return nil
}```

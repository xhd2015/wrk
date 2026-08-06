# Scenario

**Feature**: bare `wrk --force` rejects without `--push`

```
wrk --force
  -> non-zero
  -> wrk: -f/--force is only valid with --push
```

## Steps

1. Run `wrk --force` from isolated WorkRoot (L2 in-process).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--force"}
	return nil
}
```

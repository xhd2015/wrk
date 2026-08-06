# Scenario

**Feature**: bare `wrk -f` rejects without `--push`

```
wrk -f
  -> non-zero
  -> wrk: -f/--force is only valid with --push
```

## Steps

1. Run `wrk -f` from isolated WorkRoot (L2 in-process).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-f"}
	return nil
}
```

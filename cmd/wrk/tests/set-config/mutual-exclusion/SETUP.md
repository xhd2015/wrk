# Scenario

**Feature**: `--set-config` is mutually exclusive with other wrk modes

```
wrk --set-config ... --list | with create dir positional | --no-config
  -> non-zero; mutual exclusion / unexpected arguments
```

## Steps

- Leaves combine set-config with forbidden modes/args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```

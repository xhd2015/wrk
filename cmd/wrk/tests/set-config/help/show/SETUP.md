# Scenario

**Feature**: help with `--show` (and not `--create`) prints show-level usage

```
wrk --set-config --show -h|--help
  -> show usage (pretty-print config.json; missing → {})
  -> exit 0; help body is not dumped config JSON
```

## Steps

- Leaves set `--show` plus help form.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	// Level: show action help under --set-config --show.
	return nil
}
```

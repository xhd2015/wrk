# Scenario

**Feature**: `wrk -h` does not document removed `--dep` / `--all-deps` modes

```
wrk -h -> exit 0; help mentions --bring; no --dep / --all-deps mode lines
```

## Steps

1. Any cwd.
2. Run `wrk -h`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Help does not require a git consumer; use WorkRoot as cwd.
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-h"}
	return nil
}
```

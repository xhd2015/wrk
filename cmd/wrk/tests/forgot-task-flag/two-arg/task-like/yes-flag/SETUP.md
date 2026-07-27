# Scenario

**Feature**: `-y` auto-promotes task-like second positional without interactive prompt

```
wrk <dir> "fix the login bug" -y
  -> auto-promote to --task; WRK_HOME create; no confirm env required
```

## Steps

- Leaves set `Args` to include `-y` (or `--yes`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = append(req.Args, "-y")
	return nil
}
```

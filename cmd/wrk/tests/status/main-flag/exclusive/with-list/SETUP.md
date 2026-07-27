# Scenario

**Feature**: wrk --main --status --list is mutually exclusive

```
wrk --main --status --list -> non-zero; stderr mutually exclusive; empty stdout
```

## Steps

1. Parent created main repo; cwd = main root.
2. Args = `--main`, `--status`, `--list`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setMainStatusArgs(req, "--main", "--status", "--list")
	return nil
}
```
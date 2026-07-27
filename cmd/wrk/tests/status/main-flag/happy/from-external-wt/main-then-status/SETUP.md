# Scenario

**Feature**: wrk --main --status from external wt (main flag first)

```
external wt cwd -> wrk --main --status -> status of main
```

## Steps

1. Parent created main + external wt; cwd = external.
2. Args = `--main`, `--status`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setMainStatusArgs(req, "--main", "--status")
	return nil
}
```
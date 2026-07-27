# Scenario

**Feature**: wrk --main with --list is mutually exclusive

```
wrk --main --list -> non-zero; mutually exclusive
```

## Steps

1. Parent created main repo; cwd = main root.
2. Args = `--main`, `--list`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--main", "--list"}
	req.TargetDir = ""
	return nil
}
```

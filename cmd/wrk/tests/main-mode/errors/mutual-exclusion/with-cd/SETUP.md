# Scenario

**Feature**: wrk --main with --cd is mutually exclusive

```
wrk --main --cd <path> -> non-zero; mutually exclusive
```

## Steps

1. Parent created main repo; cwd = main root.
2. Args = `--main`, `--cd`, main path (path may exist; rejection is mode selection).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--main", "--cd", req.MainRepo}
	req.TargetDir = ""
	return nil
}
```

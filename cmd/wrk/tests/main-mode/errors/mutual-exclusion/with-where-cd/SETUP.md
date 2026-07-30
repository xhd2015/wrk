# Scenario

**Feature**: wrk --main --where --cd is mutually exclusive (two partners)

```
wrk --main --where --cd -> non-zero; mutually exclusive
```

## Steps

1. Parent created main repo; cwd = main root.
2. Args = `--main`, `--where`, `--cd` (no path; rejection is mode selection).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--main", "--where", "--cd"}
	req.TargetDir = ""
	return nil
}
```

# Scenario

**Feature**: wrk -h documents manual -m/--message commit message flags

```
workspace/ -> wrk -h
  -> exit 0
  -> help mentions -m and/or --message, --commit, exclusive with gen-commit-msg
```

## Steps

1. Run `wrk -h` from neutral cwd (no git required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

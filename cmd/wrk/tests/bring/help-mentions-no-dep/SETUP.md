# Scenario

**Feature**: wrk -h documents --no-dep and -v go mod tidy logging

```
wrk -h
  -> exit 0
  -> help mentions --no-dep (worktree-only / skip tidy wording)
  -> -v / --verbose help mentions go mod tidy (pre-line + child stream)
```

## Steps

1. Run `wrk -h` from isolated WorkRoot (InProcess).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

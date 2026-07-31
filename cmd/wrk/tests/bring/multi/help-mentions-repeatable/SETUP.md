# Scenario

**Feature**: `wrk -h` documents that `--bring` may be repeated for multiple deps

```
wrk -h
  -> exit 0
  -> help for --bring mentions multiple paths and/or repeatable flag form
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

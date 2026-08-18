# Scenario

**Feature**: wrk -h documents unwind/stack consumer set and --dry-run for --dep-update

```
wrk -h
  -> exit 0
  -> help mentions unwind or stack (or equivalent) for --dep-update
  -> help mentions --dry-run
```

## Steps

1. Run `wrk -h` from neutral WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"-h"}
	return nil
}
```

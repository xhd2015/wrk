# Scenario

**Feature**: `wrk -h` documents dashboard and `--new`

```
wrk -h
  -> exit 0
  -> help mentions dashboard (bare no-args helper) and --new create entry
```

## Steps

1. Run `wrk -h` from isolated WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```

# Scenario

**Feature**: bare --no-cache is only valid with --scan-git-repos

```
wrk --no-cache
  -> non-zero exit
  -> stderr: --no-cache is only valid with --scan-git-repos
  -> stdout empty
```

## Steps

1. Run `wrk --no-cache` from isolated WorkRoot (no scan mode).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--no-cache"}
	return nil
}
```

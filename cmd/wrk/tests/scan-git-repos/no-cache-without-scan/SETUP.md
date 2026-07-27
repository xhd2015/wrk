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
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--no-cache"}
	return nil
}
```

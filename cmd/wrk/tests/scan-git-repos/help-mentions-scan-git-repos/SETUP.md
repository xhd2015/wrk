# Scenario

**Feature**: wrk -h documents --scan-git-repos, --no-cache, and --include-worktrees

```
wrk -h
  -> exit 0
  -> help text contains --scan-git-repos, --no-cache, --include-worktrees
```

## Steps

1. Run `wrk -h` from isolated WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```

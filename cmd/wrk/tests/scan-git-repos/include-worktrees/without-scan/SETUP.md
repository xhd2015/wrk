# Scenario

**Feature**: bare --include-worktrees is only valid with --scan-git-repos

```
wrk --include-worktrees
  -> non-zero exit
  -> stderr: --include-worktrees is only valid with --scan-git-repos
  -> stdout empty
```

## Steps

1. Run `wrk --include-worktrees` from isolated WorkRoot (no scan mode).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--include-worktrees"}
	return nil
}
```

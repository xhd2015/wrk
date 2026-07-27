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
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--include-worktrees"}
	return nil
}
```

# Scenario

**Feature**: wrk --status --fetch on main checkout vs linked worktree cwd

```
main repo cwd + --fetch -> fetch upstream, Remote: on root block
linked wt cwd + --fetch -> silently ignored, no Remote: anywhere
```

## Steps

- Descendants set `req.Args` with `--status` and optional `--fetch`/`-v`.
- Default `req.RepoDir` is main repo unless overridden to linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureFetchVerboseHelpersUsed()
	return nil
}
```
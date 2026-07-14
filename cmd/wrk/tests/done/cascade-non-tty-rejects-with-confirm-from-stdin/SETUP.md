# Scenario

**Feature**: option A — `--confirm-from-stdin` cannot confirm cascaded ahead worktrees on non-TTY

```
consumer wt + ahead external dep -> wrk --done --confirm-from-stdin (non-TTY, stdin Y)
  -> rejected before mutations; external wt + commits preserved
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Run `wrk --done --confirm-from-stdin` with piped `\n` (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```

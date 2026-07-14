# Scenario

**Regression**: option A blocks non-TTY `--done` when cascaded external dep is ahead

```
# external dep wt ahead of dep main; --confirm-from-stdin on non-TTY must NOT merge cascade
consumer wt + ahead external/dep wt -> wrk --done --confirm-from-stdin (\n) -> error; no cascade mutations
```

## Steps

1. Create consumer main repo with `go.mod` requiring `example.com/dep`.
2. `wrk` creates the consumer linked worktree.
3. `wrk --dep <depRepo>` spawns `external/mydep-main-{date}`.
4. Commit a dep fix on the external worktree (ahead of dep main).
5. Run `wrk --done --confirm-from-stdin` from the consumer worktree, piping `\n` on non-TTY stdin.

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```
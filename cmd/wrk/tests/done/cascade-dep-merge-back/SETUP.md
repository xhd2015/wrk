# Scenario

**Feature**: non-TTY `--done` cascade-merges ahead external dep (auto-yes; no pre-flight error)

```
# external dep wt ahead of dep main; default auto-yes merges cascade then consumer
consumer wt + ahead external/dep wt -> wrk --done -> exit 0; dep merged; both wts removed
```

## Steps

1. Create consumer main repo with `go.mod` requiring `example.com/dep`.
2. `wrk` creates the consumer linked worktree.
3. `wrk --dep <depRepo>` spawns `external/mydep-main-{date}`.
4. Commit a dep fix on the external worktree (ahead of dep main).
5. Drop consumer replace/require and commit.
6. Run `wrk --done` from the consumer worktree on non-TTY.

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	prepareAheadExternalDepConsumerForDone(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```

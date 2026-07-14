# Scenario

**Feature**: `--exec` after successful `--done` runs in the main repo (not the removed worktree)

```
linked wt (already included or -y confirmed)
  -> wrk --done [-y] --exec pwd
  -> merge-back --rm; print result.Message
  -> child cmd.Dir = main repo (MergeBack TargetPath)
  -> last stdout line = main abs path
```

## Preconditions

- Linked worktree of a wrk-created branch; clean; successful done (not aborted).
- Dir is main repo never the removed worktree path.

## Steps

- Setup creates main + linked wt; leaves set `--done` + `--exec`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```


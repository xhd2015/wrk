# Scenario

**Feature**: wrk optional first positional directory argument

```
# wrk [dir] [flags...] — effective cwd is <dir> when present, else process cwd
wrk <repoDir> from WorkRoot -> same as wrk from <repoDir>
wrk <repoDir> --list from WorkRoot -> git worktree list for <repoDir>
wrk <nonexistent> -> non-zero, directory does not exist
```

## Preconditions

- Git must be available for create/list leaves.

## Steps

- Tests invoke `wrk` with `req.TargetDir` as the first positional argument when set.
- Process cwd (`req.RepoDir`) stays `{WorkRoot}` — not the target repo — to prove the dir arg sets effective cwd.
- Remaining flags pass through `req.Args` unchanged.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
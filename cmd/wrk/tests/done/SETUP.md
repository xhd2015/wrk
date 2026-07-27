# Scenario

**Feature**: wrk --done merge-back from linked worktree

```
# linked worktree created via wrk; cwd inside wt (or subpath)
wrk --done [--confirm-from-stdin] -> merge-back --rm via worktree.MergeBack
```

## Preconditions

- Git must be available.

## Steps

- Tests create a main repo, run `wrk` (no args) to add a linked worktree under `WRK_HOME`, then invoke `wrk --done` via the doctest `Run` function with `req.Args` and optional `req.StdinInput`.
- `req.RepoDir` is the cwd for `--done` (worktree root or nested subpath).
- Piped confirmation follows the mvd merge-back pattern: set `req.StdinInput` and `--confirm-from-stdin`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
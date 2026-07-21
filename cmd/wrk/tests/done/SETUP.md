# Scenario

**Feature**: wrk --done merge-back from linked worktree

```
# linked worktree created via wrk; cwd inside wt (or subpath)
# default auto-yes for [Y/n]; --confirm re-enables prompts; --confirm-from-stdin only when prompting
wrk --done [--confirm [--confirm-from-stdin]] -> merge-back --rm via worktree.MergeBack
```

## Preconditions

- Git must be available.

## Steps

- Tests create a main repo, run `wrk` (no args) to add a linked worktree under `WRK_HOME`, then invoke `wrk --done` via the doctest `Run` function with `req.Args` and optional `req.StdinInput`.
- `req.RepoDir` is the cwd for `--done` (worktree root or nested subpath).
- Decline path: `req.Args` includes `--confirm --confirm-from-stdin` and `req.StdinInput = "n\n"`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```

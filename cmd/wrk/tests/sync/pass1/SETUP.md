# Scenario

**Feature**: Pass 1 harvest — main ← linked worktree branches (FF-only)

```
# for each linked named-branch wt:
#   if main dirty / not strictly FF-ahead / WIP in main..wt → skip (+warning when applicable)
#   else git -C main merge --ff-only <wtBranch>
main + linked wt -> wrk --sync -> pass1 harvest or skip
```

## Preconditions

- Cwd is the main checkout on branch `main` (named branch).
- Linked worktrees are created with `git worktree add -b` under `{WorkRoot}/`.
- Merge refs are **named branches only** (never SHA).

## Steps

- Descendants build the main+worktree fixture and set `req.Args` / SHA fields.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```

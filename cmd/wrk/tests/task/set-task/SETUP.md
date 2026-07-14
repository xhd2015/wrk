# Scenario

**Feature**: wrk --set-task renames a linked worktree via git worktree move

```
# inside linked worktree, parse branch name, compute new slug, warn+move if TTY
wrk --set-task "new desc" -> parse wrk-shaped {branchBase}-{YYYY-MM-DD}[-slug][-N]
                         -> sanitized token for both path and new branch (no /)
                         -> suffix walk on path or branch collision
                         -> git worktree move + git branch -m (TTY required)
```

## Preconditions

- Must be run from inside a linked worktree (`.git` is a file).
- Worktree **directory name** must be wrk-shaped (contains date pattern). Fixed
  user paths / non-wrk dir names → error (unsupported).
- Branch must match wrk naming; legacy slash branches migrate to sanitized form.

## New / behavior-change leaves

- `path-collision-suffix/` — target path occupied → suffix walk (P3 T2).
- `branch-collision-suffix/` — target branch ref exists → suffix walk (P3 T3).
- `fixed-path-unsupported/` — fixed spawn path dir name → error (P3 T4).
- `legacy-slash-migrate/` — `feature/foo-{date}` → sanitized rename (P3 T5).

## Steps

- Create a worktree (with or without --task).
- Run `wrk --set-task <desc>` from inside the worktree.
- Non-TTY environment → error; TTY → confirm then move.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

```
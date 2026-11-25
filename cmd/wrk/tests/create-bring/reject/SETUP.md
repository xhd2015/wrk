# Scenario

**Feature**: create selectors without a `<project>` positional are rejected (no silent create)

```
# regression: wrk --bring d --no-new-terminal used to CREATE a worktree,
# then fail at agent launch; the worktree was already on disk.
src (git) -> wrk --bring <d1> --no-new-terminal
  -> non-zero before any work
  -> stderr names --no-new-terminal and asks for a <project> argument
  -> no new WT under WRK_HOME/worktrees; no external/; no new branch
```

## Steps

1. Create `src` (git Go module requiring dep1) + a valid `mydep1`.
2. Run the leaf's flags from `src` with **no** project positional.
3. Every leaf asserts the same shape: non-zero, flag named in stderr, zero side effects.

- This node is prose only (no `func Setup`); each leaf below carries its own fixtures + args.

# Scenario

**Feature**: create+bring failure points: preflight (no WT) vs apply (keep WT)

```
# preflightResolveBringArgs runs before create
wrk --new --bring no-such-basename -> non-zero; no new WT
# apply fail after create: keep worktree; no rollback
wrk --new --bring valid not-a-git  -> non-zero; create path exists
```

## Steps

- Fail leaves use `--no-config` and assert either no WT or kept WT without rollback.

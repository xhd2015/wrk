# Scenario

**Feature**: create+bring failure points: preflight (no WT) vs apply (keep WT)

```
# preflightResolveBringArgs runs before create
wrk <src> --new --bring no-such-basename -> non-zero; no new WT
# apply fail after create: keep worktree; no rollback
wrk <src> --new --bring valid not-a-git  -> non-zero; create path exists
```

## Steps

- Fail leaves use `--no-config` and assert either no WT or kept WT without rollback.
- Create+bring requires a `<project>` positional; leaves pass `req.TargetDir`.

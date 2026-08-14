# Scenario

**Feature**: `wrk --new --bring …` from inside `src` creates then brings

```
# explicit create from cwd; --bring applies inside the new default WT
src (cwd) -> wrk --new --no-config --bring d1
  -> {WRK_HOME}/worktrees/src-main-{date}
  -> event command=create; args include --new and --bring
```

## Steps

- `req.RepoDir = src` (no TargetDir).
- Leaves pass `--new --no-config --bring …`.

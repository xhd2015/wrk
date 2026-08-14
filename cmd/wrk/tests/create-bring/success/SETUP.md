# Scenario

**Feature**: create + `--bring` succeeds; bring consumer is the new (or spawned) worktree

```
# create form + --bring → command=create; bring into the created/reused path
src + deps -> wrk <create-form> --bring <deps> --no-config
  -> create path + external paths on stdout
  -> source src untouched
```

## Steps

- Success leaves build `src` + dep repo(s), pass `--no-config`, and assert create + bring outcomes.

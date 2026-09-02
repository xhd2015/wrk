# Scenario

**Feature**: create+`--bring` with `--here` or `--no-new-terminal` keeps the invoking terminal quiet

```
# stay-in-terminal create UX + multi --bring
#   -> create path only on stdout (no external path lines)
#   -> no "will bring:" plan; no soft SKIP lines
#   -> -v/--verbose treated as unset
#   -> externals still materialize under the new WT
src + deps -> wrk src --no-config (--here|--no-new-terminal) --bring d1 d2
```

## Preconditions

- Same fixtures as other create-bring success leaves (`--no-config`, L2 InProcess).
- Leaves under this group assert the quiet stdout/stderr contract.

## Steps

- Shared: `src` + dep repos; leave sets the quiet flag (`--here` or `--no-new-terminal`).

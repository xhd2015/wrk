# Scenario

**Feature**: multi-bring CLI rejections (exec, missing value, exact duplicates)

```
# multi + --exec          -> non-zero; --exec only with single exclusive --bring
# --bring / --bring --no-dep -> non-zero; requires a value
# --bring p1 --bring p1   -> non-zero; exact duplicate resolved path rejected
```

## Steps

- Leaves set `req.InProcess = true` and assert non-zero + stable stderr substrings.
- Success fixtures are only built when needed to prove no false GREEN (e.g. valid deps for duplicate/exec).

## Context

- Preferred exec error: `wrk: --exec is only valid with a single --bring path`
- Preferred missing value: library wording `requires a value`
- Preferred duplicate: error naming duplicate / already listed / same path (soft wording).
- `wrk --bring p1 p2` is **success** (see `../varargs-two/`), not a reject.

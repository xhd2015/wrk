# Scenario

**Feature**: `--bring` remains exclusive with non-create modes

```
# create compose does not open --done / --list / pipeline stages
wrk --done --bring d1 -> mutually exclusive
```

## Steps

- One leaf is enough; exclusive multi+`--exec` stays under `bring/multi/reject/exec`.

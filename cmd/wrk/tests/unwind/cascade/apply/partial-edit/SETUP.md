# Scenario

**Feature**: partial-edit variants when cascade pins with dirty go.mod and no `--add-all`

```
# save WT go.mod/go.sum → pin+tidy on Base → selective commit → restore + surgical bumps
# variants: unrelated WIP file; sequential pins; tidy fail restores exact WIP
dirty go.mod WIP (no --add-all) + cascade pin
  -> success paths: WIP preserved; pin commit Base-only; selective files
  -> fail path: restore snapshot; non-zero; no half-mutated WT
```

## Preconditions

- Parent apply helpers: `dirtyRootGoModWIP`, `assertPartialEditWTPreserved`,
  `assertPinCommitBaseNoWIP`, three-module / tidy-fail setup helpers.
- Primary happy path also covered by **C-AP5** (`dirty-gomod/without-add-all`).
  This grouping holds **additional** MECE scenarios (P3-2..P3-4).
- **Classic TDD RED** until partial edit lands (product still hard-Errors dirty Base).
- Do **not** rewrite clean/ or dry-run sealed leaves.

## Steps

1. Grouping scopes partial-edit edge scenarios under apply cascade.
2. Leaves seed dirty WIP (+ optional second free module / missing proxy / extra file).

## Context

- P3-5 (`--add-all` still works) remains at `dirty-gomod/with-add-all` (C-AP6) —
  no duplicate leaf here.
- Surgical bump = update require versions for cascade-pinned modules only; no
  `go mod tidy` on the restored WIP worktree.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```

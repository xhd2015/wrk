# Scenario

**Feature**: dirty go.mod WIP **with** `--add-all` still isolates pin (C-AP6 / F1)

```
# same WIP go.mod dirt as C-AP5; --add-all set but pin uses partial-edit
root (dirty go.mod vs Base)
  -> wrk --unwind --tag-next --push --add-all
  -> cascade pin on Base only (WIP restored after pin)
  -> commit "wrk: cascade pin …"; tags; exit 0; WT still dirty with WIP
```

## Steps

1. Seed apply single-repo two-module stack.
2. Append uncommitted WIP line to root go.mod.
3. Run with `--add-all` (must not disable pin WIP isolation).
4. Expect success: pin commit without WIP, WT preserves WIP + surgical bump, tags.

## Context

- **F1 (2026-08-11):** `--add-all` is a gen-commit staging flag; cascade pin always
  partial-edits when go.mod/go.sum dirty (same as without `--add-all`).
- Older C-AP6 assumed pin-on-dirty would leave go.mod clean; that scooped WIP
  into pin commits and is no longer the product contract.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoTwoModules(t, req)
	dirtyRootGoModWIP(t, req)
	// --add-all present: pin isolation must still hold (partial-edit).
	req.Args = []string{"--unwind", "--tag-next", "--push", "--add-all"}
	return nil
}
```

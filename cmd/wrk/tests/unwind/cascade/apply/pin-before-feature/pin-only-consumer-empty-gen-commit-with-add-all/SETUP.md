# Scenario

**Bug**: B1 apply — pin-only root consumer + `--add-all` empties index after
pinReady then **hard-fails** gen-commit (P-empty / production repro)

```
# mid dirty (FEATURE_WIP) + leaf clean + root staged go.mod replaces only
# (no FEATURE_WIP on root — pinReady/cascade consume all root dirt)
root-linked ← mid dirty ← leaf clean
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push --sync
  -> mid peel: feature gen-commit + land OK
  -> cascade / pinReady drop external replaces (pin auto-commits)
  -> root peel: empty index after pin → soft-skip gen-commit (even with --add-all)
  -> root merge-back; branch not diverged; exit 0
```

## Steps

1. Seed P-empty fixture (`setupApplyPinOnlyConsumerEmptyGenCommitWithAddAll`):
   clean leaf, dirty mid with FEATURE_WIP, root **go.mod-only** staged replaces
   (no root FEATURE_WIP), modproxy, bare origins.
2. Run apply with gen-commit + `--add-all` + land + `--tag-next --push --sync`.
3. Assert exit 0, mid feature landed, root pins clean (no external replace),
   pin commits never scoop replaces, root branch not diverged from main.

## Context

- Mirrors `/tmp` experimental repro of production failure:
  `no staged changes to generate commit message for` on peel `.` after pinReady
  with `--add-all` → hard abort → `Master: diverged`.
- **Today RED:**
  1. `--add-all` disables cascade partial-edit → pin subjects may scoop WIP
     external replaces mid-cascade; dirt flip lands later pins on **main** while
     earlier pins stay on the **branch**.
  2. Empty gen-commit after pinReady is a **hard error** when `--add-all` is set
     (soft-skip only without `--add-all`) → root never merge-backs.
- Desired: empty gen-commit after successful pinReady is success path when
  porcelain is clean; `--add-all` must not disable pin WIP isolation.
- Sealed T1/T2/T-M1/T-tag1 ASSERT meanings unchanged.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinOnlyConsumerEmptyGenCommitWithAddAll(t, req)
	// Flag set matches user failure path (incl. --sync like production unwind).
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
		"--sync",
	)
	return nil
}
```

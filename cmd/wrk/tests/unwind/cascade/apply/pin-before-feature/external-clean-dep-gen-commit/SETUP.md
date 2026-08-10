# Scenario

**Feature**: B1 apply — pin external clean dep **before** consumer feature gen-commit (T1)

```
# linked consumer dirty: FEATURE_WIP + go.mod replace => external/dot-pkgs (clean @ v0.0.1)
# pre-commit rejects external local replace (git-hook-go-no-local-replace sim)
root-linked ← clean leaf external
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next
  -> pin auto-commit first: wrk: cascade pin example.com/dot-pkgs @ v0.0.1
  -> then feature gen-commit (no replace; hook OK)
  -> final go.mod: require @ v0.0.1; no external replace; exit 0
```

## Steps

1. Seed T1 fixture (`setupApplyPinBeforeFeatureExternalCleanDep`): clean free dep,
   linked dirty consumer with external replace + feature WIP + pre-commit hook.
2. Run apply with gen-commit + `--add-all` + `--merge-back` + `--tag-next`.
3. Assert pin-before-feature history, require keep-current, replace dropped.

## Context

- Mirrors P1 dry-run C-DR7 shape, but **apply** + gen-commit (not dry-run).
- **Today RED:** peel/gen-commit runs with replace present → pre-commit fails
  (or, without hook, wrong history order / late pin).
- After P1 planner fix, dry-run plans the pin; apply still peels first → RED remains.
- D3: pin version = current require `v0.0.1` (no free tag).
- D7: separate pin auto-commit, then feature gen-commit.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureExternalCleanDep(t, req)
	// Flag set matches user failure path: gen-commit + land + cascade tag-next.
	// --add-all stages FEATURE_WIP for gen-commit; no --push (clean free → no
	// residual pending edges).
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
	)
	return nil
}
```

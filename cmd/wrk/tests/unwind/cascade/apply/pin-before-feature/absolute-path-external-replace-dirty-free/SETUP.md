# Scenario

**Feature**: B1 apply — absolute-path external replace + dirty free (D1)

```
# free leaf dirty → tag-next v0.0.2; consumer replace => /abs/path/to/free (not ./external)
# FEATURE_WIP staged; pre-commit forbids absolute external replace too
dirty free + abs-path replace + FEATURE_WIP
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free-first peel/tag; cascade pin drops absolute replace
  -> feature gen-commit after pin; require free @ v0.0.2; exit 0
```

## Steps

1. Seed D1 fixture (`setupApplyPinBeforeFeatureAbsolutePathReplace`): T2-like
   free-dirty stack with **absolute** replace NewPath to free checkout + FEATURE_WIP
   + extended no-local-replace hook + modproxy.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert absolute replace dropped, pin-before-feature, require@next, exit 0.

## Context

- Production wrk often uses absolute replace targets for external free checkouts.
- Product treats abs filesystem replaces as droppable when cross-repo
  (`isDroppableExternalStackReplace`); pin must drop them like `./external/…`.
- Shared fixture hook greps `./external/`, `../`, and absolute `/` so gen-commit
  cannot land while abs replace remains.
- Distinct from T2 (relative `./external/…` replace).
- P2 coverage backfill; may be GREEN immediately.
- Do not rewrite sealed T1/T2/T-M1/T-tag1/T-spl ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureAbsolutePathReplace(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

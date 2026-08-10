# Scenario

**Feature**: B1 apply — free dirty peel/tag → pin → consumer feature gen-commit (T2)

```
# free leaf dirty (owned-changed → v0.0.2) + linked consumer: external replace + FEATURE_WIP
# pre-commit rejects external local replace on consumer
leaf ← root-linked
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free-first: peel/land free → tag leaf @ v0.0.2
  -> pin auto-commit on consumer @ v0.0.2 (drop external replace)
  -> then consumer feature gen-commit (hook OK)
  -> require bumped to v0.0.2; exit 0
```

## Steps

1. Seed T2 fixture (`setupApplyPinBeforeFeatureFreeDirty`): dirty free leaf +
   dirty linked consumer with replace + feature WIP + hook + modproxy + origins.
2. Run apply with gen-commit + land + `--tag-next --push` (cross-repo pending edges).
3. Assert free tag, pin-before-feature, require bump to next.

## Context

- Free-first interleaved (D2): free peel/tag/pin complete before consumer
  feature gen-commit.
- **Today RED:** after free peel, consumer peel still gen-commits with replace
  present → hook fail; cascade pin never runs (or runs only after wrong order).
- Cross-repo residual edges → `--push` required by ValidateUnwindFlags.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFreeDirty(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

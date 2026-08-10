# Scenario

**Feature**: B1 apply resume — free already landed clean/untagged + consumer replace (A5)

```
# free main: feature commit already on main, clean porcelain, past LatestTag, no next tag
# consumer dirty: committed replace → free main + DIRTY (replace-only resume)
free clean/untagged on main + consumer replace dirt
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> no free peel (clean); cascade tags free @ v0.0.2 then pins consumer
  -> drop replace; require free @ v0.0.2; exit 0 (idempotent tag)
```

## Steps

1. Seed A5 fixture (`setupApplyPinBeforeFeatureResumeFreeLandedUntagged`):
   free main already at feature tip (clean, untagged next) + consumer dirty with
   relative replace to free main + hook + modproxy + origins.
2. Run full apply flag set (same as T2) to resume after free land.
3. Assert free tagged once @ next, consumer pinned, replace dropped, exit 0.

## Context

- Models interrupted apply after free peel/land succeeded but cascade tag/pin
  did not finish. Re-run must tag free + pin consumer without wrong double tags.
- Distinct from T2 (free still dirty linked WT needing land) and A4 (free dirty,
  consumer clean).
- P2 coverage backfill; may be GREEN immediately.
- Do not rewrite sealed T1/T2/T-M1/T-tag1/T-spl ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureResumeFreeLandedUntagged(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

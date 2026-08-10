# Scenario

**Feature**: B1 apply — clean consumer porcelain + committed replace + dirty free (A4)

```
# free leaf dirty (owned-changed → v0.0.2); consumer go.mod replace committed; WT clean
# consumer NOT in dirty peel order; cascade pin drops replace on consumer
leaf dirty ← clean consumer (committed replace only)
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> peel free only (no consumer peel before cascade)
  -> cascade: tag free @ v0.0.2; pin consumer ← free @ v0.0.2 (drop replace)
  -> require free @ v0.0.2; exit 0; no gen-commit failure on consumer
```

## Steps

1. Seed A4 fixture (`setupApplyPinBeforeFeatureCleanConsumerCommittedReplace`):
   dirty free + clean linked consumer with committed external replace (no DIRTY,
   no FEATURE_WIP) + hook + modproxy + origins.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert free tag, cascade pin drops replace, require@next, exit 0.

## Context

- A4: consumer **clean porcelain** so only free is peeled; pin still runs in
  cascade (needs-pin from droppable replace) without consumer gen-commit path.
- Distinct from T2 (consumer dirty + FEATURE_WIP) and T-spl (monorepo + DIRTY).
- P2 coverage backfill after P1; may be GREEN immediately.
- Do not rewrite sealed T1/T2/T-M1/T-tag1/T-spl ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureCleanConsumerCommittedReplace(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

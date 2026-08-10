# Scenario

**Bug**: B1 apply — false freeHost from noise intra pins peels monorepo early with dirty free replace (T-spl / A1)

```
# monorepo consumer: pkgs/shared noise pin @ LatestTag (no tag-next) + free replace
# free external dirty → tag-next v0.0.2; consumer replace-only dirt (no FEATURE_WIP)
# pre-commit rejects external local replace on consumer
dirty free + monorepo pin-consumer (noise intra pins)
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> early peel free only (monorepo must NOT freeHost from intra noise pins)
  -> cascade: tag-next free @ v0.0.2; pin consumer ← free @ v0.0.2 (drop replace)
  -> deferred consumer: soft-skip / clean; require free @ v0.0.2; exit 0
```

## Steps

1. Seed T-spl fixture (`setupApplyPinBeforeFeatureFalseFreeHostIntraPins`):
   dirty free leaf + monorepo with noise intra pin (shared require drift, no
   owned-change) + committed external replace + DIRTY-only dirt + hook + modproxy.
2. Run apply with gen-commit + land + `--tag-next --push` (same flag set as T2).
3. Assert free-first tag/pin, free advanced, consumer require@next, no hook fail.

## Context

- Encodes production **A1 / T-spl**: `splitPeelOrderB1` marks freeHost for every
  cascade pin dep, including noise intra pins @ LatestTag without tag-next →
  monorepo peels early with unready free replace → gen-commit hits
  no-local-replace hook (today RED).
- Distinct from T-M1 (true freeHost via **owned-changed** shared) and T2 (pure
  multi-repo + FEATURE_WIP). Here: false freeHost + **replace-only** consumer.
- After fix: freeHost only true tag hosts; monorepo defers; cascade pin drops
  replace before any consumer gen-commit; no feature commit that stages replace.
- Sealed T1/T2/T-M1/T-tag1 ASSERT meaning unchanged.
- Classic TDD **RED** until freeHost rule ignores non-tag pin deps.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFalseFreeHostIntraPins(t, req)
	// Flag set matches T2 free-dirty path: gen-commit + land + tag-next + push
	// (cross-repo residual edges after free tag).
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

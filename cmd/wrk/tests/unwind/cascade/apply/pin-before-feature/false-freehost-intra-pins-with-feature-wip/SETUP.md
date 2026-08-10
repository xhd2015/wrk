# Scenario

**Bug**: B1 apply — false freeHost monorepo + dirty free + FEATURE_WIP (A2)

```
# monorepo: noise intra pins (no tag-next) + free replace + staged FEATURE_WIP
# free external dirty → tag-next v0.0.2; pre-commit rejects external local replace
dirty free + monorepo pin-consumer (noise intra) + FEATURE_WIP
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> early peel free only (not false freeHost)
  -> cascade: tag free @ v0.0.2; pin consumer ← free @ v0.0.2 (drop replace)
  -> deferred consumer: feature gen-commit after pin; FEATURE_WIP lands; exit 0
```

## Steps

1. Seed A2 fixture (`setupApplyPinBeforeFeatureFalseFreeHostIntraPinsWithFeatureWIP`):
   T-spl monorepo false freeHost stack + staged FEATURE_WIP for gen-commit.
2. Run apply with gen-commit + land + `--tag-next --push` (same flag set as T2/T-spl).
3. Assert free-first tag/pin, pin ancestor of feature, FEATURE_WIP landed, require@next.

## Context

- Extends T-spl / A1 with consumer **FEATURE_WIP** (A2). freeHost must still be
  true tag hosts only so monorepo defers; cascade pin drops replace before
  feature gen-commit (D7).
- Distinct from T-spl replace-only (no feature commit required) and T2 (pure
  multi-repo, no monorepo noise intra pins).
- P2 coverage backfill: may be GREEN after P1 freeHost fix.
- Do not rewrite sealed T1/T2/T-M1/T-tag1/T-spl ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFalseFreeHostIntraPinsWithFeatureWIP(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

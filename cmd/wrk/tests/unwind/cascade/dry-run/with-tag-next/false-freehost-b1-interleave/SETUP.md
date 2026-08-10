# Scenario

**Feature**: G1 dry-run B1 order for false freeHost / T-spl-like graph

```
# monorepo: noise intra pin (shared, no tag-next) + free replace + dirty free
# B1 apply order: early free peel → cascade tag free + pin consumer → deferred consumer peel
false freeHost stack (T-spl shape)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel free (early)
  -> would: tag-next free @ v0.0.2; would: pin consumer ← free @ v0.0.2
  -> would: peel consumer (deferred after cascade)
  -> exit 0; zero mutations
```

## Steps

1. Seed G1 fixture (`setupCascadeFalseFreeHostIntraPins`): T-spl-like monorepo
   + noise intra shared pin + dirty free external replace.
2. Run dry-run with `--tag-next --push --done` (linked free WT under consumer).
3. Assert **intended** B1 interleaved dry-run order (early / cascade / deferred).

## Context

- Mirrors A1/T-spl graph used by pin-before-feature apply leaves; dry-run only.
- **Intended user-facing order** matches B1 apply (`splitPeelOrderB1`): free
  peels early, cascade tags free and pins consumer, pure pin-consumer deferred.
- Today product `FormatUnwindDryRun` may still print **all peels then cascade**
  (sealed C-DR* vocabulary). This leaf encodes the B1 interleave; **RED** until
  dry-run path reflects early/cascade/deferred (implementer fixes FormatUnwindDryRun).
- Do not rewrite sealed C-DR1–C-DR8 ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeFalseFreeHostIntraPins(t, req)
	// Linked free under consumer → --done; cross-repo residual edges → --push.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```

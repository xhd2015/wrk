# Scenario

**Bug / gap**: B1 apply — diamond stack A←B, A←C, C←B; all dirty; consumer A must get next tag (A-root-tag)

```
# production-like: checkout A with external/B + external/C
# A requires B and C; C requires B; all three dirty owned-changed → next v0.0.2
# full recipe: peel free-first, cascade pin/tag, deferred consumer feature
A (root) + B (leaf free) + C (mid freeHost)
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free-first: peel B then C (freeHost) → cascade tag/pin → deferred peel A
  -> B @ v0.0.2, C @ v0.0.2, **A @ v0.0.2 at main HEAD** after feature land
  -> exit 0
```

## Steps

1. Seed diamond fixture (`setupApplyPinBeforeFeatureDiamondAllDirtyConsumerTag`):
   dirty free B + dirty mid C (requires B) + dirty root A (requires B+C);
   owned-changed on all three; feature WIP on A; modproxy + bare origins.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert free tags, pins, and **consumer root next tag at main HEAD**.

## Context

- **Coverage gap:** multi-repo B1 leaves (T2, C-AP2, T-tag1) lock free tags +
  consumer **pins** / feature gen-commit but **never assert consumer/root A
  receives tag-next** after the full interleaved recipe.
- Product hole candidate: B1 runs cascade (pin+tag) **before** deferred consumer
  peels; feature gen-commit after cascade can leave A past LatestTag with **no**
  re-plan tag, or cascade never plans root NextTag when dirt is only pin+WIP.
- Distinct from T-tag1 (order free-tag before mid-pin; no top tag assert) and
  from A-clean-tag (free skip when clean @ LatestTag).
- L2: core flags; no `--sync`/`--reinstall-local` (tag contract first).
- Classic TDD: expect **RED** until multi-repo consumer root is tagged at tip.
- **Observed RED (pre-fix):** exit 0; free/mid ship + pins OK; root **had**
  `v0.0.2` but tag tip ≠ main HEAD after deferred feature (tagged mid-cascade).
- **GREEN after fix:** pure pin-consumer `TagNext` deferred until after feature
  peels; A next tag at main HEAD.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureDiamondAllDirtyConsumerTag(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

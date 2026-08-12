# Scenario

**Bug / gap**: B1 apply — free dirty + pure pin-consumer HEAD still at LatestTag with only uncommitted WIP must get next tag at main HEAD (A-wip-tag)

```
# production-like agent-pro hole:
# free leaf dirty (owned-changed → v0.0.2)
# consumer linked: HEAD == v0.0.1 (LatestTag); replace + FEATURE_WIP uncommitted only
# cascade plans before deferred consumer peel → tagscope sees same-commit / no NextTag
# after deferred feature land consumer tip advances untagged (crime scene REPRODUCED)
leaf ← root-linked (HEAD@v0.0.1 + WIP only)
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free @ v0.0.2; pin consumer @ v0.0.2; feature land
  -> **consumer root @ v0.0.2 at main HEAD**
  -> exit 0
```

## Steps

1. Seed fixture (`setupApplyPinBeforeFeatureFreeDirtyConsumerAtLatestWIP`):
   dirty free leaf + linked consumer whose HEAD still equals LatestTag with
   uncommitted replace + FEATURE_WIP only; modproxy + bare origins.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert free next tag and **consumer next tag at main HEAD** after full B1.

## Context

- **Coverage gap vs A-root-tag / T2:** diamond + T2 seed **commits** owned
  change or replace past LatestTag before cascade, so tagscope plans NextTag.
  Production agent-pro often runs unwind with HEAD already at the last release
  and only porcelain feature dirt → cascade plan has empty NextTag; deferred
  `applyDeferredCascadeTags` reuses the plan (no re-tagscope) → tip untagged.
- Crime-scene transcript:
  `~/.sandbox/transcripts/2026-08-12-crime-scene-unwind-agent-pro-no-tag-next.md`
- Distinct from A-root-tag (diamond mid freeHost + committed owned-changed on A)
  and T2 (consumer replace committed past tag; assert pins/feature, not root tag).
- Classic TDD: expect **RED** until product re-plans/applies consumer next tag
  after deferred feature peel when tip advanced past LatestTag.
- L2: core flags; no `--sync`/`--reinstall-local`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFreeDirtyConsumerAtLatestWIP(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

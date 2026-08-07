# Scenario

**Feature**: `--unwind --dry-run` expands stack inventory by following **local
filesystem replaces** (BFS fixpoint) into sibling/out-of-tree git checkouts

```
# seed: primary + nested under primary (unchanged)
# expand: local replace NewPath (./ | ../ | abs, no version)
#   → resolve vs module dir → ShowToplevel → extra-repo stack member + synthetic edge
# intra-repo replace → never separate member
# missing/non-git → warning: on stderr; continue
# dirty-only peel (v1); free-first among residual edges including synthetic
local filesystem replace graph
  -> wrk --unwind --dry-run [--tag-next --push when edges]
  -> would: peel <display-path>… | warning: …
```

## Preconditions

- Root helpers: `setupFollowSiblingBothDirty`, `setupFollowNestedModuleOwnsReplace`,
  `setupFollowIntraRepoOnly`, `setupFollowCleanDepSkipped`,
  `setupFollowTransitiveChain`, `setupFollowMissingTarget`,
  `setupFollowNestedModTargetToplevel`, `setPeelOrderDisplays`,
  `assertPeelOrder`, `assertUnwindZeroMutations`.
- Leaves set `req.InProcess = true` and full `req.Args`.
- Classic TDD: follow expansion not implemented yet — leaves that require
  out-of-tree members in PeelOrder are **RED** until implementer lands.

## Steps

1. Grouping marks the follow-local-replace dry-run family.
2. Leaves seed sibling/transitive/intra/missing fixtures and assert peel plan.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	unwindEnsureHelpersUsed()
	return nil
}
```

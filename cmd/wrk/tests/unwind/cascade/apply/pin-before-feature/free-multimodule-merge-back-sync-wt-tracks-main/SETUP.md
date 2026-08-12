# Scenario

**Bug / gap**: `--unwind --sync --merge-back --tag-next` peels free multi-module,
runs peel-time sync, then cascade **pin on free main** leaves the kept free
linked worktree behind main (`Master: needs fast forward(+1 commit)`)

```
# production-like (crime scene): free go-pkgs root+cmd owned-changed + consumer
# replace free root only; keep free external WT; full ship flags incl. --sync
free multi-module ext + consumer root-only
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push --sync
  -> free peel: gen-commit + merge-back + peel-time sync
  -> cascade: tag free @ v0.0.2; pin free/cmd ← free (commit on free main); pin consumer
  -> free linked WT still present; HEAD == free main HEAD (not needs FF)
  -> exit 0
```

## Steps

1. Seed C1 free multi-module fixture
   (`setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly`): dirty free root+cmd,
   droppable consumer replace free root only, modproxy, bare origins.
2. Run apply with gen-commit + `--add-all` + **`--merge-back`** + `--tag-next`
   `--push` **`--sync`** (not `--done` — free WT must remain).
3. Assert exit 0, free cascade pin landed, free linked WT kept and **tracks free
   main tip** (SHA equal; `main...branch` left-right `0 0`).

## Context

- Formalizes crime-scene transcript
  `~/.sandbox/transcripts/2026-08-12T010127Z-crime-scene-unwind-sync-dot-pkgs.md`
  (**REPRODUCED**): user free external clean + `Master: needs fast forward(+1)`.
- **Today RED:** peel-time sync only; monorepo cascade pin advances free main
  after sync; no post-cascade re-sync of free linked WTs.
- Desired: with `--sync`, free linked branch is not behind free main after full
  unwind (identical tips / zero unique commits either side).
- Sibling C1 (`free-multimodule-cmd-consumer-root-only`) locks multi-module
  tag/pin selectivity with `--done` (WT may be removed). This leaf locks
  **post-cascade sync** with **`--merge-back`** keeping the free WT.
- L2; Classic TDD RED until product re-syncs free mains after cascade pin.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly(t, req)
	// Match production unwind flags that left free WT behind main after pin.
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
		"--sync",
	)
	return nil
}
```

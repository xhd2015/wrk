# Scenario

**Feature**: clean external free already at LatestTag → no free commit, no free next tag (A-clean-tag)

```
# linked consumer dirty: FEATURE_WIP + go.mod replace => external/dot-pkgs
# free external linked WT: clean porcelain, HEAD == tag v0.0.1 (no owned-changed)
root-linked ← clean leaf external @ v0.0.1
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next
  -> pin consumer @ v0.0.1 (D3 keep-current); feature gen-commit on consumer
  -> free: no peel, no gen-commit, no tag-next / no v0.0.2
  -> exit 0
```

## Steps

1. Seed T1 fixture (`setupApplyPinBeforeFeatureExternalCleanDep`): clean free
   external already tagged at HEAD; dirty linked consumer with external replace
   + feature WIP + no-local-replace hook.
2. Run apply with gen-commit + `--add-all` + `--merge-back` + `--tag-next`
   (core flags; no `--push` — free needs no publish when untagged).
3. Assert free remains at `v0.0.1` only (no next tag / no free commit / no free peel).

## Context

- Phase A lock for: "when external linked worktree is clean (and already
  tagged at LatestTag), unwind must not create a new free tag or free commit."
- Distinct from A5 (clean free **untagged** owned-changed → **must** tag once).
- Reuses T1 seed; T1 ASSERT stays sealed (does not hard-fail free next tag).
- **Observed (Phase A run):** leaf **GREEN** immediately — product already skips
  free next tag / free peel / free HEAD move on this graph. Keep as regression
  lock; live bug (if any) needs a different fixture (e.g. owned-changed clean
  free, re-run noise, or full --push/--sync/--reinstall-local path).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureExternalCleanDep(t, req)
	// Core recipe without --push (no bare free origin required for skip contract).
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
	)
	// Baseline free tip for post-run equality (HEAD must stay at LatestTag).
	if req.SecondRepo != "" {
		writeFile(t, filepath.Join(req.WorkRoot, "_free_head.sha"),
			revParseRef(t, req.SecondRepo, "HEAD")+"\n")
	}
	return nil
}
```

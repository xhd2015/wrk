# Scenario

**Feature**: deferred free monorepo tag must re-pin stack consumers (crime-scene CS-repin)

```
# free go-pkgs monorepo: root + cmd; baseline tags; free/cmd requires free @ v0.0.0 (drift)
# **uncommitted** free root feature on WT → NextTag empty; free/cmd pin forces free deferred
# (pinConsumer without freeHost TagNext)
# peel auto-commit lands free feature; applyDeferredCascadeTags creates free @ v0.0.2
# desired: consumer require free root @ v0.0.2 (re-pin after deferred tag)
free multi-module uncommitted feature + consumer replace
  -> wrk --unwind --done --tag-next --push
  -> free tagged @ v0.0.2; consumer require free @ v0.0.2; replace dropped; exit 0
```

## Steps

1. Seed CS-repin fixture (`setupApplyDeferredTagRepinAfterFreeUncommittedFeature`):
   free monorepo root+cmd at baseline tags; free WT has **uncommitted** root
   owned change; consumer requires free root @ v0.0.1 + droppable external
   replace; offline modproxy old+next; bare origins.
2. Run apply with land + `--tag-next --push` (no gen-commit / no fake-opencode —
   free peel uses product auto-commit of dirty porcelain under `--done`).
3. Assert free next tag exists and consumer require free @ next (re-pin after
   deferred free tag).

## Context

- Formalizes crime scene: wrk tagged `go-pkgs/v0.0.120` after deferred free peel
  but left ai-critic on `v0.0.119` (no consumer re-pin).
- Distinct from C1 (free feature **committed** before apply → NextTag set →
  freeHost early peel + pin @ next in same cascade plan).
- Distinct from A5 (free already on main, clean/untagged, no monorepo cmd).
- Classic **RED** until product re-pins consumers after deferred tags (or
  treats uncommitted free monorepo as freeHost before cascade pins @ Latest).
- Do not rewrite sealed C1/T2/A5 ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyDeferredTagRepinAfterFreeUncommittedFeature(t, req)
	// No gen-commit / fake-opencode: --done auto-commits free dirty porcelain
	// during deferred free peel (crime-scene hole does not depend on AI commit).
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```

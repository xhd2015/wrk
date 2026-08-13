# Scenario

**Bug**: B1 apply — cascade intra require-drift pin tidies Base go.mod (no WIP replace) while WT source imports an unpublished free package (CS-openterm2)

```
# live spl + dirty unpublished openterm2:
# free go-pkgs HEAD==LatestTag + uncommitted shell/openterm2; cmd require-drift
# consumer HEAD go.mod has no free replace; WIP replace + import openterm2
# intra trace_tool require v0.0.15 / latest tag v0.0.16
# both peels deferred (free pinConsumer without freeHost) → cascade first
dirty free unpublished pkg + WIP-only replace + intra drift
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free tagged @ v0.0.2 with openterm2 landed
  -> consumer require free @ v0.0.2; replace dropped; intra tool @ v0.0.16
  -> no "does not contain package"; exit 0
```

## Steps

1. Seed CS-openterm2 fixture (`setupApplyIntraPinTidyUnpublishedWIPImport`):
   free monorepo at LatestTag with uncommitted `shell/openterm2`; consumer
   linked WT with intra tool drift, WIP-only external replace, and WIP import
   of the unpublished package; offline proxy latest lacks the package.
2. Run apply with gen-commit + land + `--tag-next --push` (same flag core as T2).
3. Assert exit 0, free next tag + landed package, consumer require @ next,
   no missing-package tidy diagnostic.

## Context

- Formalizes crime scene
  `~/.sandbox/transcripts/2026-08-13T02:59:27Z-crime-scene-unwind-openterm2.md`.
- Distinct from CS-repin (no new unpublished import; tidy does not die) and
  T-spl (free owned-changed committed; fail is hook not `@latest` missing
  package).
- Classic TDD leaf: was **RED** until dirty replace-targets peel early and
  partial-edit tidy overlays unrelated WIP replaces; **GREEN** after that fix.
- Do not rewrite sealed T1/T2/T-spl/CS-repin ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyIntraPinTidyUnpublishedWIPImport(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

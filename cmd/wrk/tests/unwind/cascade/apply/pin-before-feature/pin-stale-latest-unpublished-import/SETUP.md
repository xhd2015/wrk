# Scenario

**Bug**: B1 apply — mid early pinReady pins landed-but-untagged free @ LatestTag while mid imports an unpublished package (CS-pin-old-tag)

```
# production: spl → skills → go-pkgs; skills imports terminal/color added on
# uncommitted (then landed) go-pkgs HEAD; published latest is v0.0.122 without it
# B1+CS-openterm2: dirty replace-target mid peels early; pinReady uses stale
# tagscope (HEAD==Latest at plan) and pins @ old tag → tidy missing-package
leaf unpublished color/ + mid import + replace + dirty root
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push
  -> free peel/land first (color committed on main)
  -> tag-next free @ v0.0.2 then pin skills <- dot-pkgs @ v0.0.2
  -> no `@latest … does not contain package`; mid require @ v0.0.2; exit 0
```

## Steps

1. Seed CS-pin-old-tag fixture (`setupApplyPinStaleLatestUnpublishedImport`):
   leaf HEAD==LatestTag + uncommitted `color/`; mid skills replace+import
   color + FEATURE_WIP; root replaces mid+leaf; file proxy latest is old
   (no color); next zip hidden from `@v/list`.
2. Run apply with gen-commit + land + `--tag-next --push`.
3. Assert exit 0, tag-before-pin @ next, no missing-package, mid require@next.

## Context

- Formalizes crime scene
  `~/.sandbox/transcripts/2026-08-13T08:24:53Z-crime-scene-unwind-pin-old-tag.md`.
- Distinct from T-tag1 (pin @ **next** / `unknown revision`) and CS-openterm2
  (intra pin overlay; not mid pinReady @ LatestTag).
- Do not rewrite sealed T1/T2/T-tag1/CS-openterm2/CS-repin ASSERT meaning.
- Classic TDD **RED** until pin after leaf land targets the **next** tag
  (or waits for cascade tag-next) instead of published LatestTag.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinStaleLatestUnpublishedImport(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
		"--push",
	)
	return nil
}
```

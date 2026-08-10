# Scenario

**Feature**: multi-repo A→B both dirty — peel free-first **and** module cascade (C-DR2)

```
# root requires leaf; both dirty; leaf owned-changed → next v0.0.2
leaf ← root (repos + modules)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel external/dot-pkgs-main-2026-06-30
  -> would: peel .
  -> would: tag-next example.com/dot-pkgs @ v0.0.2
  -> would: pin example.com/root <- example.com/dot-pkgs @ v0.0.2
  -> exit 0; zero mutations
```

## Steps

1. Seed 2-repo stack: root main + leaf external; both dirty; leaf tagged+owned change.
2. Run dry-run with pin + land flags.
3. Expect free-first peels **and** free-first module cascade (leaf before consumer).

## Context

- Cross-repo edges → `--tag-next` + `--push`; linked leaf → `--done`.
- Module cascade is **global** (not per-repo sections).
- **RED** until cascade lines appear after peels.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeMultiRepoBothDirty(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```

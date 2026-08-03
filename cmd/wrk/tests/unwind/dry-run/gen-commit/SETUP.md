# Scenario

**Feature**: unwind dry-run gen-commit staging vocabulary (`--add-all` / leave-N)

```
# dirty primary + --gen-commit-msg --commit [--add-all] --dry-run
peel plan
  -> with --add-all: would: git add -A then generate/commit
  -> without --add-all + not fully staged: leave-N then generate/commit
  -> without --add-all + fully staged only: no leave-N; generate/commit still planned
```

## Preconditions

- Parent dry-run + root helpers (`setupSingleMainDirty`, `stageAll`, `leaveLine`).
- Leaves set full `req.Args` including `--unwind --dry-run --gen-commit-msg --commit`.
- No agent-runner required: FormatUnwindDryRun owns plan text (L2 Capture).

## Steps

1. Grouping marks gen-commit dry-run staging family.
2. Leaves seed single dirty main and choose add-all / porcelain state.

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

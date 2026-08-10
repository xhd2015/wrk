# Scenario

**Feature**: `--tag-next` enables global free-module cascade dry-run lines

```
# peels first, then free-first would: tag-next / would: pin
# replace-only external (clean free + matching require) still would: pin @ current
stack + --unwind --dry-run --tag-next [+ --push/--done as validation]
  -> peel … free-first
  -> would: tag-next <mod> @ <next>   # when owned-changed / next planned
  -> would: pin <consumer> <- <dep> @ <ver>
  -> exit 0; zero mutations
```

## Preconditions

- Leaves supply pin/land flags when the stack has cross-repo edges or linked WTs.
- Cascade lines are top-level (no indent), after peels.
- Sealed C-DR1–C-DR6 asserts unchanged; C-DR7/C-DR8 cover replace-only pin policy.

## Steps

1. Grouping locks `--tag-next` presence; leaves vary stack shape / compose flags /
   replace-only pin triggers.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```

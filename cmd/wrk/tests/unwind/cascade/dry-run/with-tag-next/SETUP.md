# Scenario

**Feature**: `--tag-next` enables global free-module cascade dry-run lines

```
# peels first, then free-first would: tag-next / would: pin
stack + --unwind --dry-run --tag-next [+ --push/--done as validation]
  -> peel … free-first
  -> would: tag-next <mod> @ <next>
  -> would: pin <consumer> <- <dep> @ <ver>
  -> exit 0; zero mutations
```

## Preconditions

- Leaves supply pin/land flags when the stack has cross-repo edges or linked WTs.
- Cascade lines are top-level (no indent), after peels.

## Steps

1. Grouping locks `--tag-next` presence; leaves vary stack shape / compose flags.

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

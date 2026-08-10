# Scenario

**Feature**: apply cascade when go.mod/go.sum WIP differs from Base

```
# pin needed but consumer go.mod already dirty vs Base
dirty go.mod + cascade pin
  -> without --add-all: partial edit (P3; was hard Error in P2)
  -> with --add-all: succeed; cascade pin commit may include extra staged dirt
```

## Preconditions

- Parent apply-cascade helpers; leaves call `dirtyRootGoModWIP` after clean fixture.
- **C-AP5** without `--add-all`: product intent **D11** — normal dirty WIP →
  **partial-edit success** (not hard Error). Sealed P2 Error assert flipped in
  P3 with documented justification (orchestrator pre-approved).
- **C-AP6** with `--add-all`: stages WIP into pin path; no partial restore (P2 GREEN).

## Steps

1. Grouping locks dirty go.mod policy.
2. Leaves split on `--add-all` presence (partial-edit vs add-all staging).

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

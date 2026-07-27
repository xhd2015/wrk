# Scenario

**Feature**: successful CPR-derived origin maps gen-commit-msg Run correctly

```
# blank above = 12; synthesize CPR for last-line cursor
BlankAbove=12
  -> ParseCPR → row1 = 12 + viewLines
  -> OriginFromCPR → originY0 = 12
  -> absY = 12 + gen-commit-msg localY
  -> ResolveMouseHit → runStage == "gen-commit-msg", OriginKind known
```

## Steps

1. Set BlankAbove=12 so root Run synthesizes CPR and aims gen-commit-msg Run.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.BlankAbove = 12
	req.Buf = nil // force synthetic CPR from BlankAbove + viewLines
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	return nil
}
```

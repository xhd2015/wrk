## Expected

- No hit with `runStage == "add-changes"`.
- `gen-commit-msg` Run hit still present (disabled gate is per-stage).

## Errors

- `err` is nil.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("hitmap: %v", err)
	}
	var haveAddRun, haveGen bool
	for _, h := range resp.Hits {
		switch h.RunStage {
		case "add-changes":
			haveAddRun = true
		case "gen-commit-msg":
			haveGen = true
		}
	}
	if haveAddRun {
		t.Fatal("add-changes Run must not appear in hitmap when AddDisabled")
	}
	if !haveGen {
		t.Fatal("gen-commit-msg Run should still be present when only add-changes is disabled")
	}
}
```

## Expected

- A Run hit exists for `gen-commit-msg`.
- A left (non-run) hit shares the same local Y with empty `RunStage` and
  `X1 <= Run.X0` (left of the Run column).
- Left hit has `Focus >= 0`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("hitmap: %v", err)
	}
	var runY, runX0, runX1 int
	var haveRun bool
	for _, h := range resp.Hits {
		if h.RunStage == "gen-commit-msg" {
			haveRun = true
			runY, runX0, runX1 = h.Y0, h.X0, h.X1
			break
		}
	}
	if !haveRun {
		t.Fatal("missing gen-commit-msg Run hit")
	}
	var haveLeft bool
	for _, h := range resp.Hits {
		if h.RunStage != "" {
			continue
		}
		if h.Y0 != runY {
			continue
		}
		if h.X1 > runX0 {
			// overlaps or is to the right — not the left toggle region
			continue
		}
		if h.Focus < 0 {
			t.Fatalf("left hit at y=%d should have focus>=0, got %d", h.Y0, h.Focus)
		}
		haveLeft = true
		break
	}
	if !haveLeft {
		t.Fatalf("missing left focus hit on gen-commit-msg row (run x=[%d,%d) y=%d)", runX0, runX1, runY)
	}
}
```

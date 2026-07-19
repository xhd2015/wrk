## Expected

- Parse and origin both succeed.
- `OriginY0 == 12`.
- Resolve hits with `RunStage == "gen-commit-msg"`.
- Prefer `OriginKind == "known"` when reported.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chain: %v", err)
	}
	if !resp.ParseOK {
		t.Fatal("expected ParseOK for synthetic CPR")
	}
	if !resp.OriginOK {
		t.Fatalf("expected OriginOK; row1=%d viewLines=%d", resp.Row1, resp.ViewLines)
	}
	if resp.OriginY0 != 12 {
		t.Fatalf("originY0: want 12, got %d", resp.OriginY0)
	}
	if !resp.ResolveOK {
		t.Fatalf("expected resolve hit; absY=%d originY0=%d", resp.AimedAbsY, resp.OriginY0)
	}
	if resp.RunStage != "gen-commit-msg" {
		t.Fatalf("runStage: want gen-commit-msg, got %q", resp.RunStage)
	}
	if resp.OriginKind != "" && resp.OriginKind != "known" {
		t.Fatalf("originKind: want known (or empty), got %q", resp.OriginKind)
	}
}
```

## Expected

- Resolve hits with `resp.RunStage == "gen-commit-msg"`.
- Prefer `resp.OriginKind == "known"` when the implementer reports origin kind.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected hit; miss absY=%d originY=%d aimedLocalY=%d",
			resp.AimedAbsY, req.OriginY, resp.AimedLocalY)
	}
	if resp.RunStage != "gen-commit-msg" {
		t.Fatalf("runStage: want gen-commit-msg, got %q", resp.RunStage)
	}
	if resp.OriginKind != "" && resp.OriginKind != "known" {
		t.Fatalf("originKind: want known (or empty), got %q", resp.OriginKind)
	}
}
```

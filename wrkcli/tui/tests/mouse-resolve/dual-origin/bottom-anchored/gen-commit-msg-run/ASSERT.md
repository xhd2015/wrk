## Expected

- Resolve hits with `resp.RunStage == "gen-commit-msg"`.
- Not `"tag-next"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected hit on gen-commit-msg Run (bottom-anchored); miss absY=%d originOffset=%d aimedLocalY=%d",
			resp.AimedAbsY, req.OriginOffset, resp.AimedLocalY)
	}
	if resp.RunStage == "tag-next" {
		t.Fatalf("bottom-anchored gen-commit-msg Run must not map to tag-next (localY=%d originKind=%q)",
			resp.LocalY, resp.OriginKind)
	}
	if resp.RunStage != "gen-commit-msg" {
		t.Fatalf("runStage: want gen-commit-msg, got %q (originKind=%q)", resp.RunStage, resp.OriginKind)
	}
}
```

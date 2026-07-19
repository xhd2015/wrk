## Expected

- Resolve misses (`resp.OK == false`).
- `resp.RunStage` is empty.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resp.OK {
		t.Fatalf("expected miss below UI; got hit runStage=%q focus=%d localY=%d originKind=%q absY=%d viewLines=%d",
			resp.RunStage, resp.Focus, resp.LocalY, resp.OriginKind, resp.AimedAbsY, resp.ViewLines)
	}
	if resp.RunStage != "" {
		t.Fatalf("miss must not carry runStage; got %q", resp.RunStage)
	}
}
```

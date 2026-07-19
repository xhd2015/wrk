## Expected

- Resolve hits with `resp.RunStage == "tag-next"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected hit; miss absY=%d originY=%d", resp.AimedAbsY, req.OriginY)
	}
	if resp.RunStage != "tag-next" {
		t.Fatalf("runStage: want tag-next, got %q", resp.RunStage)
	}
}
```

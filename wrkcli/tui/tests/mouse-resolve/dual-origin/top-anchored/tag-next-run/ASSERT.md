## Expected

- Resolve hits with `resp.RunStage == "tag-next"`.
- Not `gen-commit-msg` (distinct stage).

## Errors

- `err` is nil.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected hit on tag-next Run; miss (absY=%d aimedLocalY=%d viewLines=%d)",
			resp.AimedAbsY, resp.AimedLocalY, resp.ViewLines)
	}
	if resp.RunStage != "tag-next" {
		t.Fatalf("runStage: want tag-next, got %q (originKind=%q localY=%d)",
			resp.RunStage, resp.OriginKind, resp.LocalY)
	}
}
```

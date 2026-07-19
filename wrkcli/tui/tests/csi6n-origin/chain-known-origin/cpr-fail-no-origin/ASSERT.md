## Expected

- `ParseOK` is false (empty / failed CPR read).
- `OriginOK` is false — known origin is **not** established.
- Resolve is not required to run; `ResolveOK` stays false.

## Errors

- `err` is nil (timeout / incomplete is signaled via ok flags, not Go error).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chain: %v", err)
	}
	if resp.ParseOK {
		t.Fatalf("empty CPR: want ParseOK=false, got row1=%d col1=%d", resp.Row1, resp.Col1)
	}
	if resp.OriginOK {
		t.Fatalf("failed CPR must not set OriginOK; originY0=%d", resp.OriginY0)
	}
	if resp.ResolveOK {
		t.Fatal("failed CPR path must not claim a known-origin resolve hit")
	}
}
```

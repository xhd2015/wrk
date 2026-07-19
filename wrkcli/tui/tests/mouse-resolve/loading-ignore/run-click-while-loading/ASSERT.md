## Expected

- Resolve misses even though the click lands on a valid Run cell.
- `resp.RunStage` is empty.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !req.Loading {
		t.Fatal("test misconfigured: Loading must be true")
	}
	if resp.OK {
		t.Fatalf("loading must ignore Run click; got hit runStage=%q focus=%d",
			resp.RunStage, resp.Focus)
	}
	if resp.RunStage != "" {
		t.Fatalf("loading miss must not set runStage; got %q", resp.RunStage)
	}
}
```

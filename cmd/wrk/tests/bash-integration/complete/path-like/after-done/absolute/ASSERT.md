## Expected

- Exit code 0.
- Stdout is empty (no basenames, no flags).
- Stderr is empty.

## Side Effects

- Read-only completion; no projects.json mutation; no events.jsonl.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathLikeCompleteEmpty(t, resp)
	assertNoEventsJSONL(t, resp)
}
```

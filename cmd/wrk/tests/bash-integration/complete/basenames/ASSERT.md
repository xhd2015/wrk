## Expected

- Exit code 0.
- Stdout lists `alpha` and `alphalong` one per line (sorted), trailing newline.
- `beta` excluded by prefix filter.

## Side Effects

- Read-only completion; no projects.json mutation.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertCompleteExitOK(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 2
---
alpha
alphalong

`)
	assertNoEventsJSONL(t, resp)
}
```
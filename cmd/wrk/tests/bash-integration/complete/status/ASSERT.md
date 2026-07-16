## Expected

- Exit code 0.
- Stdout is `beta` plus trailing newline.

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
version: 3
---
beta

`)
}
```
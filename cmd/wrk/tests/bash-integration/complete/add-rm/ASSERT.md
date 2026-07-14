## Expected

- Both complete invocations exit 0.
- `--add` completion with prefix `al` returns `alpha` and `alphalong`.
- `--rm` completion with prefix `be` returns `beta`.

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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}

	addOut, ok := resp.CompleteStdout["add"]
	if !ok {
		t.Fatalf("missing add completion output")
	}
	assert.Output(t, addOut, `---
version: 2
---
alpha
alphalong

`)

	rmOut, ok := resp.CompleteStdout["rm"]
	if !ok {
		t.Fatalf("missing rm completion output")
	}
	assert.Output(t, rmOut, `---
version: 2
---
beta

`)
}
```
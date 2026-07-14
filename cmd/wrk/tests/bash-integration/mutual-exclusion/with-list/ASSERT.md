## Expected

- Non-zero exit code.
- Stderr mentions bash-integration is mutually exclusive with other modes.
- Stdout is empty.
- No `events.jsonl` created.

## Errors

- `--bash-integration` cannot be combined with `--list`.

## Exit Code

- Non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
--bash-integration is mutually exclusive
</contains>`)
	assertNoEventsJSONL(t, resp)
}
```
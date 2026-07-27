## Expected

- Exit code exactly `1`.
- Stdout empty.
- Stderr is exactly `unrecognized flag: --open-in-agen` **plus a trailing newline**
  (last byte of stderr is `\n` — not merely "contains a newline somewhere").

## Errors

- Unrecognized long flag (ordinary `error`, not `ExitCodeError`).

## Exit Code

- `1`

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 1 {
		t.Fatalf("expected exit 1, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	// Explicit trailing-byte check (must not pass if \n only appears mid-message).
	if !strings.HasSuffix(resp.Stderr, "\n") {
		t.Fatalf("stderr must end with trailing newline; got %q", resp.Stderr)
	}
	// Exact body + trailing newline policy (assert.Output enforces template trailing \n).
	assert.Output(t, resp.Stderr, "unrecognized flag: --open-in-agen\n")
}
```

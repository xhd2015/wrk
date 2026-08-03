## Expected

- Non-zero exit code.
- Stderr indicates empty / whitespace / invalid commit message.

## Side Effects

- No commit.

## Errors

- Whitespace-only `-m` value is rejected.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	errText := resp.Stderr
	low := strings.ToLower(errText)
	ok := strings.Contains(low, "empty") ||
		strings.Contains(low, "blank") ||
		strings.Contains(low, "whitespace") ||
		strings.Contains(low, "invalid") ||
		strings.Contains(low, "non-empty") ||
		strings.Contains(low, "required") ||
		strings.Contains(low, "message")
	if !ok {
		t.Fatalf("stderr should indicate empty/whitespace/invalid message, got %q", errText)
	}
	if !strings.Contains(low, "message") && !strings.Contains(errText, "-m") {
		t.Fatalf("stderr should mention message or -m, got %q", errText)
	}
}
```

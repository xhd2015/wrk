## Expected

- Non-zero exit code.
- Stderr indicates empty / invalid / blank commit message (or equivalent).

## Side Effects

- No commit.

## Errors

- Empty `-m` value is rejected.

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
		strings.Contains(low, "invalid") ||
		strings.Contains(low, "non-empty") ||
		strings.Contains(low, "required") ||
		strings.Contains(low, "missing") ||
		strings.Contains(low, "message")
	if !ok {
		t.Fatalf("stderr should indicate empty/invalid message, got %q", errText)
	}
	// Prefer message-related wording so this is not a generic unknown-flag fail.
	if !strings.Contains(low, "message") && !strings.Contains(errText, "-m") {
		t.Fatalf("stderr should mention message or -m, got %q", errText)
	}
}
```

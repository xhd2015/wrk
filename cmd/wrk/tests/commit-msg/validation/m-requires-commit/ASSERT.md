## Expected

- Non-zero exit code.
- Stderr states that `-m` / `--message` requires `--commit` (mentions message flag and `--commit`).

## Side Effects

- No commit; pure flag validation.

## Errors

- `-m` without `--commit` is rejected.

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
	hasMsgFlag := strings.Contains(errText, "-m") || strings.Contains(errText, "--message")
	if !hasMsgFlag {
		t.Fatalf("stderr should mention -m or --message, got %q", errText)
	}
	if !strings.Contains(errText, "--commit") {
		t.Fatalf("stderr should mention --commit is required, got %q", errText)
	}
	low := strings.ToLower(errText)
	if !strings.Contains(low, "require") &&
		!strings.Contains(low, "required") &&
		!strings.Contains(low, "must") &&
		!strings.Contains(low, "need") &&
		!strings.Contains(low, "only valid") {
		t.Fatalf("stderr should indicate -m requires --commit (require/must/need/only valid), got %q", errText)
	}
}
```

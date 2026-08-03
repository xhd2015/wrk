## Expected

- Non-zero exit code.
- Stderr indicates `--commit` needs a message source: `-m`/`--message` and/or `--gen-commit-msg`.

## Side Effects

- No commit.

## Errors

- Bare `--commit` is incomplete without a message source.

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
	if !strings.Contains(errText, "--commit") {
		t.Fatalf("stderr should mention --commit, got %q", errText)
	}
	hasManual := strings.Contains(errText, "-m") || strings.Contains(errText, "--message")
	hasGen := strings.Contains(errText, "--gen-commit-msg") || strings.Contains(strings.ToLower(errText), "gen-commit")
	if !hasManual && !hasGen {
		t.Fatalf("stderr should point at -m/--message and/or --gen-commit-msg as message source, got %q", errText)
	}
	// Prefer both alternatives named when product can say so.
	low := strings.ToLower(errText)
	if !strings.Contains(low, "require") &&
		!strings.Contains(low, "required") &&
		!strings.Contains(low, "must") &&
		!strings.Contains(low, "need") &&
		!strings.Contains(low, "missing") &&
		!strings.Contains(low, "provide") {
		t.Fatalf("stderr should indicate a message source is needed, got %q", errText)
	}
}
```

## Expected

- Non-zero exit.
- Stderr mentions `--commit` (required with primary).
- Stderr mentions the primary `--done` and/or “primary”.
- Prefer require/required/must/need wording.

## Errors

- `--model` does not substitute for `--commit` when composing with `--done`.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --gen-commit-msg --model=m --done without --commit, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--commit") {
		t.Fatalf("stderr should mention --commit is required with primary, got %q", se)
	}
	if !strings.Contains(se, "--done") && !strings.Contains(strings.ToLower(se), "primary") {
		t.Fatalf("stderr should mention --done or primary, got %q", se)
	}
	lower := strings.ToLower(se)
	if !strings.Contains(lower, "require") &&
		!strings.Contains(lower, "required") &&
		!strings.Contains(lower, "must") &&
		!strings.Contains(lower, "need") {
		t.Fatalf("stderr should indicate --commit is required with primary (require/must/need), got %q", se)
	}
}
```

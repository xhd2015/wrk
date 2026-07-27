## Expected

- Non-zero exit.
- Stderr mentions `--dir`.
- Stderr indicates reject with primary (`--done` / not valid / cannot / mutually exclusive).
- Prefer primary-aware wording over silent accept of `--dir`.

## Errors

- `--dir` cannot be composed with `--gen-commit-msg` + `--done` (wrk workDir wins).

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --gen-commit-msg --commit --dir … --done, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--dir") {
		t.Fatalf("stderr should mention --dir, got %q", se)
	}
	if !strings.Contains(se, "--done") &&
		!strings.Contains(strings.ToLower(se), "primary") &&
		!strings.Contains(strings.ToLower(se), "not valid") &&
		!strings.Contains(strings.ToLower(se), "mutually exclusive") &&
		!strings.Contains(strings.ToLower(se), "cannot") {
		t.Fatalf("stderr should reject composed --dir with primary, got %q", se)
	}
}
```

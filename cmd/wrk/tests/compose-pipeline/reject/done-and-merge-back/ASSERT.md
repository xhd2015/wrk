## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion / cannot combine.
- Mentions `--done` and `--merge-back`.

## Errors

- Done and merge-back cannot compose.

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
		t.Fatalf("expected non-zero for --done --merge-back; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate exclusion, got %q", se)
	}
	if !strings.Contains(se, "--done") && !strings.Contains(se, "done") {
		t.Fatalf("stderr should mention done, got %q", se)
	}
	if !strings.Contains(se, "--merge-back") && !strings.Contains(se, "merge-back") {
		t.Fatalf("stderr should mention merge-back, got %q", se)
	}
}
```

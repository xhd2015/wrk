## Expected

- Non-zero exit; empty stdout.
- Stderr indicates `--exec` is not valid with `--where` (mode policy), not merely unknown flag.
- Preferred: names both `--exec` and `--where` with not-valid / mutually exclusive phrasing.

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
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)

	se := resp.Stderr
	if strings.Contains(se, "unrecognized") {
		t.Fatalf("stderr is parse-unknown only (%q); want mode policy rejecting --exec with --where", se)
	}
	if !strings.Contains(se, "--exec") {
		t.Fatalf("stderr should mention --exec, got %q", se)
	}
	if !strings.Contains(se, "--where") {
		t.Fatalf("stderr should mention --where, got %q", se)
	}
	if !strings.Contains(se, "not valid") &&
		!strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "only valid") {
		t.Fatalf("stderr should indicate not-valid/mutual-exclusion policy, got %q", se)
	}
}
```

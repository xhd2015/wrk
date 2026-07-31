## Expected

- Non-zero exit.
- Stderr indicates a mode conflict (mutually exclusive / not valid / cannot).
- Prefer mentioning both `--pr` and `--main` when product wording allows.

## Errors

- `--pr` and `--main` cannot be combined.

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
		t.Fatalf("expected non-zero for --pr --main, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "only valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate mutual exclusion / validity conflict, got %q", se)
	}
}
```

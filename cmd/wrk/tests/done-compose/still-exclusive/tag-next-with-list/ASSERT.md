## Expected

- Non-zero exit.
- Stderr mentions `mutually exclusive` (or equivalent mode conflict).
- Prefer mentioning both `--tag-next` and `--list`.

## Errors

- `--tag-next` and `--list` cannot be combined.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --tag-next --list, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate mutual exclusion, got %q", se)
	}
	if !strings.Contains(se, "--tag-next") && !strings.Contains(se, "tag-next") {
		// soft: at least one mode named; current product says "--tag-next is mutually exclusive with other modes"
		t.Fatalf("stderr should mention tag-next, got %q", se)
	}
}
```

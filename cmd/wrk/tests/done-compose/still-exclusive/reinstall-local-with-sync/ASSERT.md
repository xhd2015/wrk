## Expected

- Non-zero exit.
- Stderr mentions `mutually exclusive` (or equivalent mode conflict).
- Prefer mentioning `--reinstall-local` and/or `--sync`.

## Errors

- Bare `--reinstall-local` and `--sync` cannot be combined without a primary.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --reinstall-local --sync, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate mutual exclusion, got %q", se)
	}
	if !strings.Contains(se, "--reinstall-local") &&
		!strings.Contains(se, "reinstall-local") &&
		!strings.Contains(se, "--sync") {
		t.Fatalf("stderr should name reinstall-local and/or sync, got %q", se)
	}
}
```

## Expected

- Non-zero exit code.
- Stderr indicates mutual exclusion (or equivalent mode conflict).
- Prefer mentioning `--pin-locals` and/or `--done`.
- Stdout empty (or no successful pin-local plan).

## Errors

- `--pin-locals` cannot be combined with `--done`.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertMutualExclusion(t, resp)
	// Prefer naming at least one of the conflicting flags.
	se := resp.Stderr
	if !strings.Contains(se, "--pin-locals") &&
		!strings.Contains(se, "pin-locals") &&
		!strings.Contains(se, "--done") {
		t.Fatalf("stderr should name --pin-locals and/or --done, got %q", se)
	}
}
```

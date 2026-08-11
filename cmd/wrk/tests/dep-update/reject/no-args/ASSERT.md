## Expected

- Non-zero exit.
- Stderr indicates missing directory argument and/or need for `--all`.
- No `dep-update ` success lines.

## Errors

- Empty paths and no `--all` → Error requires directory or `--all`.

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
	assertExitNonZero(t, resp)
	se := strings.ToLower(resp.Stderr)
	// Accept pre-migration "requires a directory" and post "directory or --all".
	if !strings.Contains(se, "director") &&
		!strings.Contains(se, "path") &&
		!strings.Contains(se, "requires") &&
		!strings.Contains(se, "argument") &&
		!strings.Contains(se, "--all") {
		t.Fatalf("stderr should indicate missing directory or --all, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "dep-update ")
}
```

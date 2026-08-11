## Expected

- Non-zero exit.
- Stderr product reject: `--dep-update --all` cannot take directory/path args
  (not merely unknown-flag parse failure).
- No successful pin / tidy / summary lines.

## Errors

- `--dep-update --all` with remaining path args → error.

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
	if strings.Contains(se, "unrecognized") || strings.Contains(se, "unknown flag") {
		t.Fatalf("want product reject for --all+dirs, still unknown/unrecognized: %q", resp.Stderr)
	}
	if !strings.Contains(se, "all") {
		t.Fatalf("stderr should mention --all, got %q", resp.Stderr)
	}
	// Directory / path / argument conflict signal.
	if !strings.Contains(se, "director") &&
		!strings.Contains(se, "path") &&
		!strings.Contains(se, "arg") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "not") {
		t.Fatalf("stderr should reject --all with directory args, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertNotContains(t, resp.Stdout, "dep-update: updated")
	assertNotContains(t, resp.Stdout, "dep-update: would update")
}
```

## Expected

- Non-zero exit.
- Stderr product reject: `--all` is only valid with `--dep-update` (must name
  `dep-update`, not merely "unrecognized flag").
- No inventory pin success lines.

## Errors

- Bare `--all` is not a standalone mode.

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
	// Classic TDD: unknown-flag-only is insufficient — product must register
	// --all as a partner and reject bare use with dep-update-oriented wording.
	if strings.Contains(se, "unrecognized") || strings.Contains(se, "unknown flag") {
		t.Fatalf("want partner validation for bare --all, still unknown/unrecognized: %q", resp.Stderr)
	}
	if !strings.Contains(se, "dep-update") {
		t.Fatalf("stderr should mention --dep-update when rejecting bare --all, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "all") {
		t.Fatalf("stderr should mention --all, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "dep-update: updated")
	assertNotContains(t, resp.Stdout, "would: dep-update")
}
```

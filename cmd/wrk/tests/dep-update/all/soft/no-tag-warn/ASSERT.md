## Expected

- Exit **0** (soft skip, not hard error like dir-mode no-tags).
- Stderr has `warning:` (or warning) about missing tag/version.
- No successful `dep-update <mod> ->` pin line for the untagged module.
- Summary reflects zero updates and skipped ≥1
  (`dep-update: updated 0, already 0, skipped 1` preferred).
- Consumer go.mod unchanged.

## Errors

- Soft only — must not hard-fail like dir-mode `error/no-tags`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertWarningStderr(t, resp.Stderr)
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "tag") && !strings.Contains(se, "version") {
		t.Fatalf("warning should mention tag/version, got %q", resp.Stderr)
	}
	// No pin success for the untagged module.
	assertNoPinFor(t, resp.Stdout, modLib)
	for _, line := range strings.Split(resp.Stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "dep-update "+modLib) && strings.Contains(trim, "->") {
			t.Fatalf("must not pin untagged module: %q", trim)
		}
	}
	assertAllSummary(t, resp.Stdout, req.WantUpdated, req.WantAlready, req.WantSkipped, wantCheckoutsOf(req), false)
	assertGoModUnchanged(t, req)
	assertOwnerGoModUnchanged(t, req)
}
```

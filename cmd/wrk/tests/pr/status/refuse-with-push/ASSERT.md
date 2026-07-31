## Expected

- Non-zero exit.
- Stderr indicates invalid combination of PR status with `--push` (mutually exclusive / not valid / only valid / cannot / invalid).
- Prefer mentioning `--status` and/or `--push` when product wording allows.
- Origin `refs/heads/feature-pr` equals pre-run snapshot (no full tip push).
- No status stdout block; no `pushed` line.

## Errors

- `--pr --status` is read-only; cannot compose with `--push` (push-existing is `--pr --push` without `--status`).

## Side Effects

- Origin tip unchanged (no ensure-push / full push).

## Exit Code

- Non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --pr --status --push; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "only valid") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "invalid") {
		t.Fatalf("stderr should indicate invalid combination / exclusion, got %q", resp.Stderr)
	}
	for _, tok := range []string{"State:", "Checks:", "pushed ", "PR created"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("refuse must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}

	// Origin tip must remain at pre-run snapshot (no full push).
	beforeBytes, readErr := os.ReadFile(filepath.Join(req.WorkRoot, "origin-feature-before"))
	if readErr != nil {
		t.Fatalf("read origin snapshot: %v", readErr)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := revParseRef(t, req.OriginBare, "refs/heads/"+req.WtBranch)
	if after != before {
		t.Fatalf("origin/%s must not advance under --pr --status --push; before=%s after=%s",
			req.WtBranch, before, after)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	_ = invocs
}
```

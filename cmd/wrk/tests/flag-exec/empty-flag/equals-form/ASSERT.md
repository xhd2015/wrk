## Expected

- Non-zero exit.
- Stderr rejects **equals form** `--exec=value` (cut is not a value flag), with wording distinct from plain `unrecognized flag: --exec`.
- Preferred implementer wording (soft): mention equals form / `=` / that cut does not take `=value` (must fail on unrecognized-only).
- Stdout empty; no worktree created.

## Errors

- `--exec=value` is invalid.

## Exit Code

- Non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}

	se := resp.Stderr
	// Reject false-GREEN: current binary only says "unrecognized flag: --exec".
	if strings.Contains(se, "unrecognized") {
		t.Fatalf("stderr is parse-unknown only (%q); want equals-form rejection for --exec=value", se)
	}
	if !strings.Contains(se, "--exec") {
		t.Fatalf("stderr should mention --exec, got %q", se)
	}
	// Equals-form signal: literal '=', or "equals", or "value" / "does not take" style.
	if !strings.Contains(se, "=") &&
		!strings.Contains(strings.ToLower(se), "equals") &&
		!strings.Contains(se, "value form") &&
		!strings.Contains(se, "does not take") {
		t.Fatalf("stderr should reject equals form (contain '=' or equals/value-form wording), got %q", se)
	}

	wtRoot := filepath.Join(req.WrkHome, "worktrees")
	if entries, err := os.ReadDir(wtRoot); err == nil && len(entries) > 0 {
		t.Fatalf("no worktree should be created; worktrees=%v", entries)
	}
}
```

## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `dashboard` (case-insensitive OK).
- Help text contains `--new` as the create entry (not only `--new-window` / `--new-terminal`).
- Prefer also mentioning bare `wrk` does not create / opens dashboard.

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	lower := strings.ToLower(help)
	if !strings.Contains(lower, "dashboard") {
		t.Fatalf("help must mention dashboard; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Require standalone create flag documentation: line or token "--new" not only as --new-window prefix.
	if !strings.Contains(help, "--new") {
		t.Fatalf("help must mention --new; got %q", help)
	}
	// Prefer explicit create-entry wording near --new (flag list or usage line).
	hasCreateEntry := strings.Contains(lower, "create a worktree") ||
		strings.Contains(lower, "explicit create") ||
		strings.Contains(help, "wrk --new")
	if !hasCreateEntry {
		t.Fatalf("help should document --new as create entry; got %q", help)
	}
}
```

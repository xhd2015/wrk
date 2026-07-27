
## Expected

- Exit code 0.
- Stdout is non-empty and ends with trailing `\n`.
- Stdout registers compspec with default filename fallback:
  `complete -o default -F _wrk wrk`.
- Stdout `_wrk` path-like branch yields via `compopt -o default`.
- Stdout contains `WRK_HOME` resolution for script path.
- Stdout contains `--bash-integration --complete` callback invocation.
- Stderr is empty.
- No `events.jsonl` created.

## Side Effects

- Read-only; no profile or integration file writes.

## Exit Code

- 0

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertContains(t, resp.Stdout, "complete -o default -F _wrk wrk")
	assertContains(t, resp.Stdout, "compopt -o default")
	assertContains(t, resp.Stdout, "WRK_HOME")
	assertContains(t, resp.Stdout, "--bash-integration --complete")
	assertNoEventsJSONL(t, resp)
	if _, statErr := os.Stat(resp.BashShPath); !os.IsNotExist(statErr) {
		t.Fatalf("print-script must not write bash.sh at %s", resp.BashShPath)
	}
}
```

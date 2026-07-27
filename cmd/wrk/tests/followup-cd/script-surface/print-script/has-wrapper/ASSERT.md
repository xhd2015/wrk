---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit code 0.
- Stdout non-empty, trailing `\n`.
- Stdout defines a `wrk()` function (wrapper, not only `_wrk` completion).
- Stdout mentions `WRK_FOLLOWUP_FILE` and `WRK_AUTO_CD`.
- Stdout still registers `complete -o default -F _wrk wrk`.
- Stdout mentions `--bash-integration --complete` callback.
- Stderr empty.

## Side Effects

- Read-only; no bash.sh write.

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
	if !scriptDefinesWrkWrapper(resp.Stdout) {
		t.Fatalf("expected wrk() wrapper function in printed script; got:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "WRK_FOLLOWUP_FILE")
	assertContains(t, resp.Stdout, "WRK_AUTO_CD")
	assertContains(t, resp.Stdout, "complete -o default -F _wrk wrk")
	assertContains(t, resp.Stdout, "--bash-integration --complete")
	if _, statErr := os.Stat(bashShPath(req.WrkHome)); !os.IsNotExist(statErr) {
		t.Fatalf("print-script must not write bash.sh at %s", bashShPath(req.WrkHome))
	}
}
```

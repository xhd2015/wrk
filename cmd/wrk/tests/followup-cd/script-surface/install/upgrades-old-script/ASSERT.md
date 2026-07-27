---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit code 0.
- bash.sh no longer is the completion-only stub.
- bash.sh defines `wrk()` and mentions `WRK_FOLLOWUP_FILE`.

## Side Effects

- Overwrites pre-seeded bash.sh content.

## Exit Code

- 0

```go
import (
	"strings"
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
	content := resp.BashShContent
	if content == "" {
		t.Fatal("bash.sh content empty after install")
	}
	if content == req.PreExistingBashSh {
		t.Fatalf("install must overwrite old bash.sh; content unchanged")
	}
	if strings.Contains(content, "pre-seeded completion-only") {
		t.Fatalf("old stub content should not remain:\n%s", content)
	}
	if !scriptDefinesWrkWrapper(content) {
		t.Fatalf("upgraded bash.sh must define wrk() wrapper; got:\n%s", content)
	}
	assertContains(t, content, "WRK_FOLLOWUP_FILE")
	assertContains(t, content, "complete -o default -F _wrk wrk")
}
```

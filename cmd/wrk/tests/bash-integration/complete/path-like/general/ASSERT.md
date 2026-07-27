---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Each path-like prefix (`./`, `../`, `/`) at positional cword 1 yields empty stdout.
- Exit code 0 for the primary case; no custom basenames leaked despite seeded projects.

## Side Effects

- Read-only completion; no events.jsonl.

## Exit Code

- 0

```go
import (
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
	for _, name := range []string{"relative-dot", "parent-relative", "absolute"} {
		out, ok := resp.CompleteStdout[name]
		if !ok {
			t.Fatalf("missing complete case %q", name)
		}
		if out != "" {
			t.Fatalf("path-like case %q must yield empty stdout; got %q", name, out)
		}
	}
	assertNoEventsJSONL(t, resp)
}
```

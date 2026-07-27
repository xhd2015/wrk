---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Basename completion with empty projects returns empty stdout (exit 0).
- Flag completion still returns flag candidates including `--list` and `--bash-integration`.

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

	basenameOut, ok := resp.CompleteStdout["basename"]
	if !ok {
		t.Fatalf("missing basename completion output")
	}
	if basenameOut != "" {
		t.Fatalf("expected empty basename stdout, got %q", basenameOut)
	}

	flagsOut, ok := resp.CompleteStdout["flags"]
	if !ok {
		t.Fatalf("missing flags completion output")
	}
	assertCompleteExitOK(t, &Response{Stdout: flagsOut, ExitCode: 0})
	assertAllLinesAreFlags(t, flagsOut)
	assertFlagsInclude(t, flagsOut, "--list", "--bash-integration")
}
```
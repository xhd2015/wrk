## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents:
  - `--pr`
  - `--title`
  - `--comment`
- Stdout (preferred for usage) ends with trailing `\n` when non-empty.

## Side Effects

- Read-only (`-h` only).

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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text for wrk -h")
	}
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout should end with trailing newline, got %q", resp.Stdout)
	}
	for _, want := range []string{"--pr", "--title", "--comment"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help must mention %q; got stdout=%q stderr=%q", want, resp.Stdout, resp.Stderr)
		}
	}
}
```

## Expected

- Exit 0 (or help-style success).
- Help text (stdout and/or stderr) contains `--dep-replace`.

## Exit Code

- 0 (help may also use non-zero on some CLIs; prefer 0 like pin-locals)

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	// Help typically exits 0 for -h.
	if resp.ExitCode != 0 {
		// Accept non-zero only if help text still present (flag.ErrHelp patterns).
		help := resp.Stdout + resp.Stderr
		if !strings.Contains(help, "--dep-replace") {
			t.Fatalf("help must mention --dep-replace; exit=%d stdout=%q stderr=%q",
				resp.ExitCode, resp.Stdout, resp.Stderr)
		}
		return
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--dep-replace") {
		t.Fatalf("help must mention --dep-replace; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```

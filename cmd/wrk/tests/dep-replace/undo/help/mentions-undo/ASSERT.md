## Expected

- Exit 0 (or help success).
- Help text contains `--undo`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--undo") {
		t.Fatalf("help must mention --undo; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 0 && !strings.Contains(help, "--dep-replace") {
		t.Fatalf("non-zero help without --dep-replace; exit=%d", resp.ExitCode)
	}
}
```
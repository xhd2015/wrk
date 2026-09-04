## Expected

- Non-zero exit.
- Error mentions `--web` and `--dry-run`.

## Exit Code

- non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	msg := resp.Stdout + resp.Stderr
	if !strings.Contains(msg, "dry-run") {
		t.Fatalf("expected dry-run in error, got %q", msg)
	}
}
```

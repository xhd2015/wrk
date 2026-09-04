## Expected

- Exit code 0.
- Help mentions `--unwind --web` and `--port`.

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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--unwind --web") {
		t.Fatalf("help must mention --unwind --web; got %q", help)
	}
	if !strings.Contains(help, "--port") {
		t.Fatalf("help must mention --port; got %q", help)
	}
}
```

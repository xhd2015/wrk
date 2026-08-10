## Expected

- Exit 0 (or help success with text).
- Help text contains `--dep-update`.

## Exit Code

- 0 preferred

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
	if !strings.Contains(help, "--dep-update") {
		t.Fatalf("help must mention --dep-update; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}
```

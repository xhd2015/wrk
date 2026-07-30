## Expected

- Exit 0.
- Stdout is exactly main abs path + `\n` (from `pwd` with cmd.Dir=main).
- Follow-up file is `cd <main>\n`.
- In-place mode stdout alone would be empty; with `--exec`, child stdout only.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolvePath(t, req.MainRepo)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))
	assertFollowupCDLine(t, req, want)
}
```

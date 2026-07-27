## Expected

- Exit code 0.
- Stdout is exactly the jump directory absolute path (from `pwd`) + trailing `\n`.
- Follow-up file is exactly `cd <abs>\n`.
- No interactive shell required (follow-up channel open).

## Side Effects

- Follow-up file written; exec ran with `cmd.Dir` = jump abs.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	// In-place --cd alone prints nothing; with --exec, child stdout only.
	assert.Output(t, resp.Stdout, v2StdoutTemplate(req.MainRepo))
	assertFollowupCDExact(t, req, req.MainRepo)
}
```

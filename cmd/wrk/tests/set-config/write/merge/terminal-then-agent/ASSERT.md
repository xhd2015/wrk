## Expected

- Exit 0.
- `terminal.mode=new` preserved.
- `agent` enabled with defaults.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertTerminalMode(t, req.WrkHome, "new")
	assertDefaultAgentOn(t, req.WrkHome)
}
```

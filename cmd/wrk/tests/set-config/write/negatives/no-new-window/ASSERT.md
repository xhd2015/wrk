## Expected

- Exit 0.
- Window absent or not mode=new.
- Terminal still present (`new`) unless implementer also clears it (prefer preserve terminal).

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
	assertWindowAbsentOrOff(t, req.WrkHome)
	// terminal should remain unless implementation documents otherwise
	assertTerminalMode(t, req.WrkHome, "new")
}
```

## Expected

- Exit 0; empty stdout.
- `window.mode=new` and `terminal.mode=new`.
- Agent not required to be present (not requested).

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertWindowModeNew(t, req.WrkHome)
	assertTerminalMode(t, req.WrkHome, "new")
	create := readCreateSection(t, req.WrkHome)
	if _, ok := create["agent"]; ok {
		// tolerate agent absent or untouched; must not force-enable without flag
		assertAgentEnabled(t, req.WrkHome, false)
	}
}
```

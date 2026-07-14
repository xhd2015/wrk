## Expected

- Exit code 0.
- Empty stdout (management mutator style).
- `create.window.mode` is `new`.
- `create.terminal.mode` is `new`.
- `create.agent` enabled with defaults: runner `grok-tty`, prompt `/brainstorm ${task}`, default args.

## Side Effects

- Writes `{WRK_HOME}/config.json`.

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
	assertDefaultAgentOn(t, req.WrkHome)
}
```

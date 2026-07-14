## Expected

- Exit 0.
- Top-level `extra` is still number `1`.
- `create.terminal.mode=new`.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertTerminalMode(t, req.WrkHome, "new")
	root := readSetConfigRoot(t, req.WrkHome)
	raw, ok := root["extra"]
	if !ok {
		t.Fatal("expected top-level extra key preserved")
	}
	var extra float64
	if err := json.Unmarshal(raw, &extra); err != nil {
		t.Fatalf("extra: %v", err)
	}
	if extra != 1 {
		t.Fatalf("extra: want 1, got %v", extra)
	}
}
```

## Expected

- Exit 0.
- Stdout is JSON containing create window/terminal/agent markers (pretty or compact).
- Non-empty stdout ending with `\n`.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected non-empty JSON stdout from --show")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout must end with newline: %q", resp.Stdout)
	}
	// Accept full config or create-section-only JSON.
	var any interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &any); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, resp.Stdout)
	}
	s := resp.Stdout
	for _, needle := range []string{"window", "terminal", "agent", "new"} {
		if !strings.Contains(s, needle) {
			t.Fatalf("show stdout missing %q:\n%s", needle, s)
		}
	}
}
```

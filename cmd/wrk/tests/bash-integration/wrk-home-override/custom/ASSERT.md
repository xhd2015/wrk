## Expected

- Exit code 0.
- `bash.sh` written under custom WRK_HOME, not default `{WorkRoot}/.wrk`.
- Both profile markers reference `$WRK_HOME` (not hardcoded `~/.wrk` only).

## Side Effects

- Script and dual profile markers written under custom WRK_HOME resolution.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}

	customWrk := filepath.Join(req.WorkRoot, "custom-wrk")
	defaultWrk := filepath.Join(req.WorkRoot, ".wrk")
	customScript := filepath.Join(customWrk, "integration", "bash.sh")
	defaultScript := filepath.Join(defaultWrk, "integration", "bash.sh")

	if _, statErr := os.Stat(customScript); statErr != nil {
		t.Fatalf("expected bash.sh under custom WRK_HOME %s: %v", customScript, statErr)
	}
	if _, statErr := os.Stat(defaultScript); !os.IsNotExist(statErr) {
		t.Fatalf("install must not write bash.sh under default WRK_HOME %s", defaultScript)
	}
	for _, content := range []string{resp.BashProfileContent, resp.BashRCContent} {
		if !strings.Contains(content, "${WRK_HOME:-$HOME/.wrk}") {
			t.Fatalf("marker must use WRK_HOME resolution:\n%s", content)
		}
		if strings.Contains(content, filepath.Join("$HOME", ".wrk")) && !strings.Contains(content, "WRK_HOME") {
			t.Fatalf("marker must not hardcode only ~/.wrk:\n%s", content)
		}
	}
	assertWrkHomeIsolated(t, customScript, customWrk)
	assertNoEventsJSONL(t, resp)
}
```
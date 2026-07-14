## Expected

- Exit code 0.
- `{WRK_HOME}/integration/bash.sh` exists.
- Script defines `wrk()` wrapper (not only `_wrk` completion).
- Script mentions `WRK_FOLLOWUP_FILE` and `WRK_AUTO_CD`.
- Script still registers `complete -o default -F _wrk wrk`.

## Side Effects

- bash.sh written under isolated WRK_HOME.

## Exit Code

- 0

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("expected bash.sh at %s: %v", resp.BashShPath, statErr)
	}
	content := resp.BashShContent
	if content == "" {
		t.Fatalf("bash.sh content empty at %s", resp.BashShPath)
	}
	if !scriptDefinesWrkWrapper(content) {
		t.Fatalf("installed bash.sh must define wrk() wrapper; got:\n%s", content)
	}
	assertContains(t, content, "WRK_FOLLOWUP_FILE")
	assertContains(t, content, "WRK_AUTO_CD")
	assertContains(t, content, "complete -o default -F _wrk wrk")
}
```

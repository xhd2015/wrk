## Expected Output

```
bash integration: installed
script: {WRK_HOME}/integration/bash.sh (installed)
bash_profile: {HOME}/.bash_profile (marker installed)
bashrc: {HOME}/.bashrc (marker installed)

```

## Expected

- Exit code 0.
- Stdout matches the four-line installed report with absolute paths and trailing blank line.
- `bash.sh` written; one marker in each profile.
- No `events.jsonl`.

## Side Effects

- Creates `integration/bash.sh`.
- Appends marker blocks to both profiles (creates files if absent).

## Exit Code

- 0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "installed", "installed", "installed", "installed")
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("expected bash.sh at %s: %v", resp.BashShPath, statErr)
	}
	if !strings.Contains(resp.BashShContent, "complete -o default -F _wrk wrk") {
		t.Fatalf("bash.sh must register complete -o default -F _wrk wrk:\n%s", resp.BashShContent)
	}
	assertMarkersInstalled(t, resp)
	assertNoEventsJSONL(t, resp)
}
```

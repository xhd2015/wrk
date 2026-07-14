## Expected

- Exit code 0.
- `{WRK_HOME}/integration/bash.sh` exists and registers `complete -o default -F _wrk wrk`
  with path-like yield (`compopt -o default`).
- `~/.bash_profile` and `~/.bashrc` each contain exactly one wrk marker block.
- Both profiles source bash.sh via `$WRK_HOME` resolution in the marker.
- No `events.jsonl` created.

## Side Effects

- `integration/bash.sh` written under WRK_HOME.
- Marker block appended to both profile files (created if absent).

## Exit Code

- 0

```go
import (
	"os"
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
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("expected bash.sh at %s: %v", resp.BashShPath, statErr)
	}
	assertContains(t, resp.BashShContent, "complete -o default -F _wrk wrk")
	assertContains(t, resp.BashShContent, "compopt -o default")
	if resp.BashProfileMarkerCount != 1 {
		t.Fatalf("expected 1 marker in .bash_profile, got %d:\n%s", resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 1 {
		t.Fatalf("expected 1 marker in .bashrc, got %d:\n%s", resp.BashRCMarkerCount, resp.BashRCContent)
	}
	for _, content := range []string{resp.BashProfileContent, resp.BashRCContent} {
		if !strings.Contains(content, "WRK_HOME") {
			t.Fatalf("profile marker must reference WRK_HOME:\n%s", content)
		}
		if !strings.Contains(content, "integration/bash.sh") {
			t.Fatalf("profile marker must source integration/bash.sh:\n%s", content)
		}
	}
	assertHomeIsolated(t, resp.BashProfilePath, resp.Home)
	assertHomeIsolated(t, resp.BashRCPath, resp.Home)
	assertWrkHomeIsolated(t, resp.BashShPath, resp.WrkHome)
	assertNoEventsJSONL(t, resp)
}
```
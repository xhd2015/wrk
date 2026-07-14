## Expected

- Exit code 0.
- Both profiles no longer contain wrk marker blocks.
- Unrelated profile content preserved.
- `integration/bash.sh` still exists with original content.
- No `events.jsonl` created.

## Side Effects

- Marker blocks removed from `.bash_profile` and `.bashrc`.
- `integration/bash.sh` not deleted.

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
	if resp.BashProfileMarkerCount != 0 {
		t.Fatalf("expected .bash_profile marker removed; count=%d:\n%s",
			resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 0 {
		t.Fatalf("expected .bashrc marker removed; count=%d:\n%s",
			resp.BashRCMarkerCount, resp.BashRCContent)
	}
	for _, content := range []string{resp.BashProfileContent, resp.BashRCContent} {
		if !strings.Contains(content, "export EDITOR=vim") {
			t.Fatalf("uninstall must preserve unrelated profile content:\n%s", content)
		}
	}
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("bash.sh must remain after uninstall: %v", statErr)
	}
	if !strings.Contains(resp.BashShContent, "pre-seeded wrk bash integration") {
		t.Fatalf("uninstall must not modify bash.sh:\n%s", resp.BashShContent)
	}
	assertNoEventsJSONL(t, resp)
}
```
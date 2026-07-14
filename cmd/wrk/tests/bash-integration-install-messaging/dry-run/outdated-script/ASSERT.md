## Expected Output

```
bash integration: would update
script: {WRK_HOME}/integration/bash.sh (would update)
bash_profile: {HOME}/.bash_profile (marker is up to date)
bashrc: {HOME}/.bashrc (marker is up to date)

```

## Expected

- Exit code 0.
- Summary is **`would update`**, not “already installed / no changes needed”.
- Script line uses `(would update)` because content differs from embedded.
- Markers report `is up to date`.
- Pre-seeded bash.sh and profiles unchanged.
- No `events.jsonl`.

## Side Effects

- None (read-only dry-run).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "would update", "would update", "is up to date", "is up to date")
	assertDryRunUnchanged(t, resp)
	if !strings.Contains(resp.BashShContent, "pre-seeded outdated wrk bash integration") {
		t.Fatalf("dry-run must leave outdated bash.sh untouched:\n%s", resp.BashShContent)
	}
	if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
		t.Fatalf("dry-run must preserve .bash_profile:\n%s", resp.BashProfileContent)
	}
	assertNoEventsJSONL(t, resp)
}
```

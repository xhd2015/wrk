
## Expected Output

```
bash integration: updated
script: {WRK_HOME}/integration/bash.sh (updated)
bash_profile: {HOME}/.bash_profile (marker is up to date)
bashrc: {HOME}/.bashrc (marker is up to date)

```

## Expected

- Exit code 0.
- Stdout reports summary `updated` and script `(updated)`.
- Both markers remain single and report `is up to date`.
- `bash.sh` content is rewritten (no longer the outdated pre-seed stub).
- No `events.jsonl`.

## Side Effects

- Overwrites outdated `bash.sh` with embedded script.
- Does not duplicate profile markers.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "updated", "updated", "is up to date", "is up to date")
	if strings.Contains(resp.BashShContent, "pre-seeded outdated wrk bash integration") {
		t.Fatalf("install must rewrite outdated bash.sh; still has stub:\n%s", resp.BashShContent)
	}
	if !strings.Contains(resp.BashShContent, "complete -o default -F _wrk wrk") {
		t.Fatalf("updated bash.sh must register complete -o default -F _wrk wrk:\n%s", resp.BashShContent)
	}
	assertMarkersInstalled(t, resp)
	if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
		t.Fatalf("must preserve unrelated .bash_profile content:\n%s", resp.BashProfileContent)
	}
	assertNoEventsJSONL(t, resp)
}
```

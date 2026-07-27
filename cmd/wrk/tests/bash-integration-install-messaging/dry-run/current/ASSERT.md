---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected Output

```
bash integration: is up to date
script: {WRK_HOME}/integration/bash.sh (is up to date)
bash_profile: {HOME}/.bash_profile (marker is up to date)
bashrc: {HOME}/.bashrc (marker is up to date)

```

## Expected

- Exit code 0.
- Stdout four-line `is up to date` report (no “already installed” / “no changes needed”).
- Filesystem unchanged from pre-install state.
- No `events.jsonl`.

## Side Effects

- None (read-only dry-run).

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "is up to date", "is up to date", "is up to date", "is up to date")
	assertDryRunUnchanged(t, resp)
	assertMarkersInstalled(t, resp)
	assertNoEventsJSONL(t, resp)
}
```

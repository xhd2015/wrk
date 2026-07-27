
## Expected Output

```
bash integration: would update
script: {WRK_HOME}/integration/bash.sh (is up to date)
bash_profile: {HOME}/.bash_profile (marker would install)
bashrc: {HOME}/.bashrc (marker would install)

```

## Expected

- Exit code 0.
- Summary `would update` (markers would be installed; script already current).
- Script line `(is up to date)`; both markers `(marker would install)`.
- No profile files created; bash.sh unchanged.
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
	assertInstallReport(t, resp, "would update", "is up to date", "would install", "would install")
	assertDryRunUnchanged(t, resp)
	if resp.BashShContent != req.PreExistingBashSh {
		t.Fatalf("dry-run must not rewrite bash.sh")
	}
	assertNoEventsJSONL(t, resp)
}
```


## Expected Output

```
bash integration: would install
script: {WRK_HOME}/integration/bash.sh (would install)
bash_profile: {HOME}/.bash_profile (marker would install)
bashrc: {HOME}/.bashrc (marker would install)

```

## Expected

- Exit code 0.
- Stdout four-line `would install` report with absolute paths and trailing blank line.
- No `bash.sh`, `.bash_profile`, or `.bashrc` created.
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
	assertInstallReport(t, resp, "would install", "would install", "would install", "would install")
	assertDryRunUnchanged(t, resp)
	assertNoEventsJSONL(t, resp)
}
```

## Expected Output

```
bash integration: updated
script: {WRK_HOME}/integration/bash.sh (is up to date)
bash_profile: {HOME}/.bash_profile (marker installed)
bashrc: {HOME}/.bashrc (marker installed)

```

## Expected

- Exit code 0.
- Summary is `updated` (markers installed; script already current).
- Script line: `(is up to date)`; both markers `(marker installed)`.
- Markers appended; script content unchanged from embedded.
- No `events.jsonl`.

## Side Effects

- Appends markers to both profiles.
- Does not rewrite current bash.sh (content stays the embedded script).

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "updated", "is up to date", "installed", "installed")
	if resp.BashShContent != req.PreExistingBashSh {
		t.Fatalf("current bash.sh must not be rewritten when already embedded-equal")
	}
	assertMarkersInstalled(t, resp)
	assertNoEventsJSONL(t, resp)
}
```

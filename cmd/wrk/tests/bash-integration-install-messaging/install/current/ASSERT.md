## Expected Output

```
bash integration: is up to date
script: {WRK_HOME}/integration/bash.sh (is up to date)
bash_profile: {HOME}/.bash_profile (marker is up to date)
bashrc: {HOME}/.bashrc (marker is up to date)

```

## Expected

- Exit code 0.
- Stdout reports all components `is up to date`.
- Exactly one marker remains in each profile.
- No `events.jsonl`.

## Side Effects

- No duplicate markers.
- Script remains present.

## Exit Code

- 0

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExit0(t, resp, err)
	assertInstallReport(t, resp, "is up to date", "is up to date", "is up to date", "is up to date")
	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("bash.sh missing after current re-install: %v", statErr)
	}
	assertMarkersInstalled(t, resp)
	assertNoEventsJSONL(t, resp)
}
```


## Expected Output

```
bash integration: updated
script: {WRK_HOME}/integration/bash.sh (is up to date)
bash_profile: {HOME}/.bash_profile (marker installed)
bashrc: {HOME}/.bashrc (marker is up to date)

```

## Expected

- Exit code 0.
- Summary `updated` because one marker was installed.
- Exactly one marker in each profile after install.
- Unrelated `.bashrc` content preserved.
- No `events.jsonl`.

## Side Effects

- Appends marker only to `.bash_profile`.
- Leaves existing `.bashrc` marker count at 1.

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
	assertInstallReport(t, resp, "updated", "is up to date", "installed", "is up to date")
	assertMarkersInstalled(t, resp)
	if !strings.Contains(resp.BashRCContent, "export EDITOR=vim") {
		t.Fatalf("must preserve unrelated .bashrc content:\n%s", resp.BashRCContent)
	}
	assertNoEventsJSONL(t, resp)
}
```

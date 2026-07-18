## Expected

- Exit code 0.
- Stdout reports `is up to date` with script and both profile marker states (current embedded script + markers).
- Pre-installed bash.sh and profile content unchanged by dry-run.

## Side Effects

- No profile modifications.
- No bash.sh overwrite.
- No `events.jsonl`.

## Exit Code

- 0

```go
import (
"fmt"
"os"
"strings"
"testing"

"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
if err != nil {
t.Fatalf("Run error: %v", err)
}
if resp.ExitCode != 0 {
t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
}

assert.Output(t, resp.Stdout, fmt.Sprintf(`---
version: 3
---
bash integration: is up to date
script: %s \(is up to date\)
bash_profile: %s \(marker is up to date\)
bashrc: %s \(marker is up to date\)

`, resp.BashShPath, resp.BashProfilePath, resp.BashRCPath))

if resp.BashProfileMarkerCount != 1 {
t.Fatalf("dry-run must not change .bash_profile marker count; got %d:\n%s",
resp.BashProfileMarkerCount, resp.BashProfileContent)
}
if resp.BashRCMarkerCount != 1 {
t.Fatalf("dry-run must not change .bashrc marker count; got %d:\n%s",
resp.BashRCMarkerCount, resp.BashRCContent)
}
if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
t.Fatalf("dry-run must preserve unrelated .bash_profile content:\n%s", resp.BashProfileContent)
}
if !strings.Contains(resp.BashShContent, "complete -o default -F _wrk wrk") {
t.Fatalf("bash.sh must still register complete -o default after dry-run:\n%s", resp.BashShContent)
}
if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
t.Fatalf("bash.sh must still exist: %v", statErr)
}
assertNoEventsJSONL(t, resp)
}
```

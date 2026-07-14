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
- No bash.sh or profile files created.
- No `events.jsonl`.

## Side Effects

- No profile modifications.
- No bash.sh write.
- No `events.jsonl`.

## Exit Code

- 0

```go
import (
"fmt"
"os"
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
version: 2
---
bash integration: would install
script: %s (would install)
bash_profile: %s (marker would install)
bashrc: %s (marker would install)

`, resp.BashShPath, resp.BashProfilePath, resp.BashRCPath))

if _, statErr := os.Stat(resp.BashShPath); !os.IsNotExist(statErr) {
t.Fatalf("dry-run must not create bash.sh at %s", resp.BashShPath)
}
if _, statErr := os.Stat(resp.BashProfilePath); !os.IsNotExist(statErr) {
t.Fatalf("dry-run must not create .bash_profile at %s", resp.BashProfilePath)
}
if _, statErr := os.Stat(resp.BashRCPath); !os.IsNotExist(statErr) {
t.Fatalf("dry-run must not create .bashrc at %s", resp.BashRCPath)
}
assertNoEventsJSONL(t, resp)
}
```

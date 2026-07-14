## Expected

- Exit code 0 after second install.
- Exactly one marker block remains in each profile.
- `bash.sh` remains present; install upgrades outdated pre-seed content to the
  embedded script (rewrite-on-diff is intentional for wrapper upgrades).
- No duplicate markers.

## Side Effects

- No duplicate markers appended.
- Outdated pre-seeded bash.sh is rewritten to the current embedded script when
  content differs (upgrade path); markers stay single.

## Exit Code

- 0

```go
import (
"os"
"strings"
"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
if err != nil {
t.Fatalf("Run error: %v", err)
}
if resp.ExitCode != 0 {
t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
}
if resp.BashProfileMarkerCount != 1 {
t.Fatalf("idempotent install must not duplicate .bash_profile marker; count=%d:\n%s",
resp.BashProfileMarkerCount, resp.BashProfileContent)
}
if resp.BashRCMarkerCount != 1 {
t.Fatalf("idempotent install must not duplicate .bashrc marker; count=%d:\n%s",
resp.BashRCMarkerCount, resp.BashRCContent)
}
if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
t.Fatalf("bash.sh missing after idempotent install: %v", statErr)
}
if !strings.Contains(resp.BashShContent, "complete -o default -F _wrk wrk") {
t.Fatalf("bash.sh must register complete -o default after install:\n%s", resp.BashShContent)
}
if !strings.Contains(resp.BashShContent, "compopt -o default") {
t.Fatalf("bash.sh must path-like yield via compopt -o default:\n%s", resp.BashShContent)
}
// Second install must not re-append markers; unrelated profile content preserved.
if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
t.Fatalf("must preserve unrelated .bash_profile content:\n%s", resp.BashProfileContent)
}
assertNoEventsJSONL(t, resp)
}
```

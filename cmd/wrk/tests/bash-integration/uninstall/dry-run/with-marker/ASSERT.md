## Expected

- Exit code 0.
- Stdout previews marker block removal from both profiles.
- Profile marker counts remain 1 in each file.

## Side Effects

- No profile modifications.
- No bash.sh delete.

## Exit Code

- 0

```go
import (
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

	assert.Output(t, resp.Stdout, `---
version: 3
---
dry-run: would remove marker block from ~/\.bash_profile
dry-run: would remove marker block from ~/\.bashrc

# === wrk integration begin ===
_wrk_home="\$\{WRK_HOME:-\$HOME/\.wrk\}"
\[\[ -f "\$_wrk_home/integration/bash\.sh" \]\] && source "\$_wrk_home/integration/bash\.sh"
# === wrk integration end ===

`)

	if resp.BashProfileMarkerCount != 1 {
		t.Fatalf("dry-run must not remove .bash_profile marker; count=%d:\n%s",
			resp.BashProfileMarkerCount, resp.BashProfileContent)
	}
	if resp.BashRCMarkerCount != 1 {
		t.Fatalf("dry-run must not remove .bashrc marker; count=%d:\n%s",
			resp.BashRCMarkerCount, resp.BashRCContent)
	}
	if !strings.Contains(resp.BashProfileContent, "export EDITOR=vim") {
		t.Fatalf("dry-run must preserve unrelated .bash_profile content:\n%s", resp.BashProfileContent)
	}
	assertHomeIsolated(t, resp.BashProfilePath, resp.Home)
	assertHomeIsolated(t, resp.BashRCPath, resp.Home)
}
```
## Expected

- Exit code 1.
- Stdout reports `partial` with script present and both profile markers absent.
- Pre-seeded bash.sh unchanged.

## Side Effects

- Read-only inspection.

## Exit Code

- 1

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
	if resp.ExitCode != 1 {
		t.Fatalf("expected exit 1, got %d; stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assert.Output(t, resp.Stdout, fmt.Sprintf(`---
version: 2
---
bash integration: partial
script: %s (present)
bash_profile: %s (marker absent)
bashrc: %s (marker absent)

`, resp.BashShPath, resp.BashProfilePath, resp.BashRCPath))

	if _, statErr := os.Stat(resp.BashShPath); statErr != nil {
		t.Fatalf("expected pre-seeded bash.sh present: %v", statErr)
	}
	if !strings.Contains(resp.BashShContent, "wrk integration stub") {
		t.Fatalf("status must not modify bash.sh:\n%s", resp.BashShContent)
	}
	if _, statErr := os.Stat(resp.BashProfilePath); !os.IsNotExist(statErr) {
		t.Fatalf("status must not create .bash_profile at %s", resp.BashProfilePath)
	}
	if _, statErr := os.Stat(resp.BashRCPath); !os.IsNotExist(statErr) {
		t.Fatalf("status must not create .bashrc at %s", resp.BashRCPath)
	}
	assertWrkHomeIsolated(t, resp.BashShPath, resp.WrkHome)
}
```
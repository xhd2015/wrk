
## Expected

- Exit code 0.
- Stdout reports already uninstalled for both profiles.
- No profile or bash.sh files created.

## Side Effects

- No filesystem writes.

## Exit Code

- 0

```go
import (
	"fmt"
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", resp.ExitCode, resp.Stderr)
	}

	assert.Output(t, resp.Stdout, fmt.Sprintf(`---
version: 3
---
wrk bash integration: already uninstalled
bash_profile: %s \(marker absent\)
bashrc: %s \(marker absent\)
no changes needed

`, resp.BashProfilePath, resp.BashRCPath))

	if _, statErr := os.Stat(resp.BashProfilePath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create .bash_profile at %s", resp.BashProfilePath)
	}
	if _, statErr := os.Stat(resp.BashRCPath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create .bashrc at %s", resp.BashRCPath)
	}
	if _, statErr := os.Stat(resp.BashShPath); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create bash.sh at %s", resp.BashShPath)
	}
}
```
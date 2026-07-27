## Expected

- Non-zero exit.
- `--set-config` is recognized (not "unrecognized flag").
- No new worktree under `{WRK_HOME}/worktrees/`.
- Stderr indicates unexpected arguments / mutual exclusion / set-config conflict.

## Errors

- `--set-config` is not create mode and does not accept create positionals.

## Exit Code

- non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when combining create dir with set-config, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected set-config/dir conflict after --set-config is implemented; got %q", resp.Stderr)
	}
	wtRoot := filepath.Join(req.WrkHome, "worktrees")
	entries, err := os.ReadDir(wtRoot)
	if err == nil && len(entries) > 0 {
		t.Fatalf("set-config must not create worktrees, found %v under %s", entries, wtRoot)
	}
	if resp.Stderr == "" {
		t.Fatal("expected stderr explaining rejection")
	}
}
```

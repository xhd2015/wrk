## Expected

- Non-zero exit.
- Both `--no-config` and `--set-config` are recognized (not "unrecognized flag").
- Stderr names the mutual exclusion (preferred exact body prefix):
  `wrk: --no-config is mutually exclusive with --set-config`
- No `config.json` created under `{WRK_HOME}`.
- Empty stdout preferred.

## Errors

- Management mode cannot run alongside `--no-config` (config skip flag).

## Side Effects

- No config write; no worktree under `{WRK_HOME}/worktrees/`.

## Exit Code

- non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --no-config + --set-config, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected mutual-exclusion after --no-config is implemented; got %q", resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--no-config") {
		t.Fatalf("stderr should mention --no-config: %q", se)
	}
	if !strings.Contains(se, "--set-config") {
		t.Fatalf("stderr should mention --set-config: %q", se)
	}
	if !(strings.Contains(se, "mutually exclusive") || strings.Contains(se, "exclusive")) {
		assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
	}
	cfg := setConfigPath(req.WrkHome)
	if _, err := os.Stat(cfg); err == nil {
		t.Fatalf("mutual exclusion must not write config.json; found %s", cfg)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat config.json: %v", err)
	}
	wtRoot := filepath.Join(req.WrkHome, "worktrees")
	entries, err := os.ReadDir(wtRoot)
	if err == nil && len(entries) > 0 {
		t.Fatalf("must not create worktrees, found %v under %s", entries, wtRoot)
	}
}
```

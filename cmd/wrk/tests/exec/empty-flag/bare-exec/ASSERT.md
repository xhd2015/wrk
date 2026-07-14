## Expected

- Non-zero exit.
- Stderr indicates `--exec` requires a command / arguments.
- Stdout empty; no worktree under WRK_HOME.

## Errors

- Cut marker with no trailing args.

## Exit Code

- Non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
--exec
</contains>`)
	// Prefer "requires" wording; soft-match command/argument synonyms.
	se := resp.Stderr
	if !strings.Contains(se, "requires") && !strings.Contains(se, "command") && !strings.Contains(se, "argument") {
		t.Fatalf("stderr should mention requires/command/argument, got %q", se)
	}

	wtRoot := filepath.Join(req.WrkHome, "worktrees")
	if entries, err := os.ReadDir(wtRoot); err == nil && len(entries) > 0 {
		t.Fatalf("no worktree should be created; worktrees=%v", entries)
	}
}
```

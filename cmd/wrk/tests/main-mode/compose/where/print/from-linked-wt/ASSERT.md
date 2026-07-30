## Expected

- Exit code 0.
- Stdout is exactly the **main repo** absolute path + trailing `\n` (not the linked worktree path).
- Stderr is empty.
- Fake interactive shell was **not** launched.

## Side Effects

- No nested shell; no follow-up write.

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutMainPath(t, resp.Stdout, req.MainRepo)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertFakeShellNotLaunched(t, req)
	if resolvePath(t, req.WtDir) == resolvePath(t, req.MainRepo) {
		t.Fatal("fixture error: linked worktree path equals main repo")
	}
	// Must print main, not linked path.
	if strings.TrimSpace(resp.Stdout) == resolvePath(t, req.WtDir) {
		t.Fatalf("stdout must be main path, not linked wt; got %q", resp.Stdout)
	}
}
```

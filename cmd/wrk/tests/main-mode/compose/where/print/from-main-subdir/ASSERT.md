## Expected

- Exit 0.
- Stdout is main repo root abs path + `\n` (not the subdir).
- Empty stderr; no shell.

## Exit Code

- 0

```go
import (
	"path/filepath"
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
	if strings.Contains(strings.TrimSpace(resp.Stdout), string(filepath.Separator)+"pkg"+string(filepath.Separator)) {
		t.Fatalf("stdout must be main root, not a subpath; got %q", resp.Stdout)
	}
}
```

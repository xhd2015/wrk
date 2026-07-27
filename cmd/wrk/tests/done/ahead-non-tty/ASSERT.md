## Expected

- Exit code 0.
- Worktree directory removed; branch deleted.
- Main repo contains the worktree commit.
- No `Proceed?` prompt required.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	combined := strings.ToLower(resp.Stdout + resp.Stderr)
	if strings.Contains(combined, "proceed?") {
		t.Fatalf("default auto-yes must not prompt, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```

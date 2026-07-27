## Expected

- Exit code 0 (aborted cleanly, same as plain decline).
- Stdout contains `merge-back aborted`.
- Stdout does **not** contain `synced:`.
- Worktree still exists; main does not have `feature-work`; branch still present.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "merge-back aborted")
	assertNotContains(t, resp.Stdout, "synced:")

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```

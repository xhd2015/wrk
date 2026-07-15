## Expected

- Exit code 0.
- Primary merge message, blank line, zero-summary sync.
- No `worktree removed:`.
- wtA still exists; branch exists; main has `feature-work`.

## Exit Code

- 0

```go
import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)
	assertNotContains(t, resp.Stdout, "worktree removed:")

	primary := fmt.Sprintf("merged branch %s into main", req.WtBranch)
	want := primaryThenSyncStdout(primary, nil, 0, 0, 0)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
}
```

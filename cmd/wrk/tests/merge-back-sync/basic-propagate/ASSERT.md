## Expected

- Exit code 0.
- Stdout: `merged branch <WtBranch> into main`, blank line, pass2 for wtB, summary.
- No `worktree removed:` in stdout.
- wtA still exists with branch; main has `feature-work`.
- wtB HEAD equals main HEAD.

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
	want := primaryThenSyncStdout(primary, []string{syncDetailPass2(req.Wt2Branch, 1)}, 0, 1, 0)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertHEAD(t, req.Wt2Dir, revParseHEAD(t, req.MainRepo))
}
```

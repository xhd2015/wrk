## Expected Output

Primary merge message, blank line, pass-2 detail for `feature-stays`, blank line, summary.

## Expected

- Exit code 0.
- Stdout: `merged branch <WtBranch> into main`, blank line, then sync detail+summary for wtB.
- Stderr empty.
- wtA directory gone; branch deleted; main has `feature-work`.
- wtB HEAD equals main HEAD after distribute.

## Exit Code

- 0

```go
import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	primary := fmt.Sprintf("merged branch %s into main", req.WtBranch)
	want := primaryThenSyncStdout(primary, []string{syncDetailPass2(req.Wt2Branch, 1)}, 0, 1, 0)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))

	mainSHA := revParseHEAD(t, req.MainRepo)
	assertHEAD(t, req.Wt2Dir, mainSHA)
}
```

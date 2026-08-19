## Expected Output

```
merged branch <WtBranch> into main

pushed main → origin/main
pushed <WtBranch> → origin/<WtBranch>
```

## Expected

- Exit code 0.
- Stderr empty.
- Worktree and local branch gone.
- `origin/main` and `origin/<WtBranch>` both equal post-land main HEAD.

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
	want := primaryThenTwoPushStdout(primary, req.WtBranch)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	mainSHA := revParseHEAD(t, req.MainRepo)
	assertOriginBranchEquals(t, req.OriginBare, "main", mainSHA)
	assertOriginBranchEquals(t, req.OriginBare, req.WtBranch, mainSHA)
}
```

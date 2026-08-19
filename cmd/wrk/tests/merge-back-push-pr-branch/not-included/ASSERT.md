## Expected

- Exit code 0.
- Stdout: primary merge + `pushed main → origin/main` only (no same-name push line).
- Stderr `warning:` that the origin tip is not in the local branch.
- `origin/<WtBranch>` still at the remote-only commit.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	primary := fmt.Sprintf("merged branch %s into main", req.WtBranch)
	want := primary + "\n\n" + "pushed main → origin/main\n"
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))
	if strings.Contains(resp.Stdout, fmt.Sprintf("pushed %s → origin/%s", req.WtBranch, req.WtBranch)) {
		t.Fatalf("must not push origin/%s when tip is not in local; stdout=%q", req.WtBranch, resp.Stdout)
	}

	if !strings.Contains(resp.Stderr, "warning:") ||
		!strings.Contains(resp.Stderr, "not in local branch") {
		t.Fatalf("stderr should warn about remote-only commits, got %q", resp.Stderr)
	}

	saved := loadSavedOriginBranchTip(t, req)
	assertOriginBranchEquals(t, req.OriginBare, req.WtBranch, saved)
	mainSHA := revParseHEAD(t, req.MainRepo)
	assertOriginBranchEquals(t, req.OriginBare, "main", mainSHA)
}
```

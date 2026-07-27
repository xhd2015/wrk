## Expected

- Exit code 0.
- Root block `Dir: .` includes `Remote:       identical`.
- Linked worktree block has `Master:` but **no** `Remote:`.
- Stderr is empty.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := strings.Count(resp.Stdout, "Remote:"); got != 1 {
		t.Fatalf("expected exactly one Remote: line, got %d:\n%s", got, resp.Stdout)
	}

	root := statusRootBlockWithRemotePlain(t, req.MainRepo, "clean", "Remote:       identical")
	master := masterIdenticalField(t, req.MainRepo, "main", "wt-ident")
	wt := fmt.Sprintf("Dir:          wt-linked\n%s\n%s\nStatus:       clean\n%s",
		statusBranchLine(t, req.WtDir), statusCommitLine(t, req.WtDir), master)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(joinStdoutBlocks(root, wt)))
}
```
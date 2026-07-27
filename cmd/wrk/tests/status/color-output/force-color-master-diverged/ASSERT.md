## Expected

- Exit code 0.
- Linked worktree `Master:` value `diverged(2 commits)` is wrapped in red ANSI.
- Stderr is empty.

## Exit Code

- 0

```go
import (
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
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	master := colorStatusMasterFieldColored(t, req.MainRepo, "main", req.WtBranch)
	assert.Output(t, resp.Stdout, colorStatusStdoutV2(t,
		colorStatusBlockPlain(t, req.MainRepo, ".", "<ansi-color green>clean</ansi-color>", ""),
		colorStatusBlockPlain(t, req.WtDir, "wt-linked", "<ansi-color green>clean</ansi-color>", master),
	))
}
```
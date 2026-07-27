## Expected

- Exit code 0.
- `Worktrees:` shows `3 total, ` uncolored and `1 dirty` in red ANSI.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"fmt"

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
	assertColorProjectsBlocksSeparated(t, resp.Stdout, 1)

	wantSummary := colorLinkedWorktreeSummary(t, req.MainRepo)
	if wantSummary != "3 total, 1 dirty" {
		t.Fatalf("helper summary: want 3 total, 1 dirty, got %q", wantSummary)
	}
	remote := colorCompareWithRemoteField(t, req.MainRepo, "origin/main", "main")

	block := fmt.Sprintf(`---
version: 3
---
%s
%s
%s
Status:       clean
%s
Worktrees:    3 total, <ansi-color red>1 dirty</ansi-color>
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo), remote)
	assert.Output(t, resp.Stdout, block)
}
```
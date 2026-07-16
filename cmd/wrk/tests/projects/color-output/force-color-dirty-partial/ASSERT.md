## Expected

- Exit code 0.
- `Status:` uses granular coloring: red for `dirty` and `2 changed`, grey (`#90`) for zero-count segments.
- Separators `(`, `, `, `)` are uncolored.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"fmt"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertColorProjectsBlocksSeparated(t, resp.Stdout, 1)

	remote := colorCompareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	status := colorFormatDirtyStatusCounts(0, 2, 0, 0)
	block := fmt.Sprintf(`---
version: 3
---
%s
%s
%s
Status:       %s
%s
Worktrees:    0 total, 0 dirty
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo), status, remote)
	assert.Output(t, resp.Stdout, block)
}
```
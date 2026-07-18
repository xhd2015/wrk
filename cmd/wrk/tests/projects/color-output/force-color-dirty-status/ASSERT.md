## Expected

- Exit code 0.
- `Status:` value uses granular coloring when `--color` is set: red for `dirty` and non-zero count segments, grey (`#90`) for zero-count segments.
- Other fields (including `Worktrees:`) are not erroneously colored.
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
	block := fmt.Sprintf(`---
version: 3
---
%s
%s
%s
Status:       <ansi-color red>dirty</ansi-color> \(<ansi-color red>1 added</ansi-color>, <ansi-color #90>0 changed</ansi-color>, <ansi-color #90>0 renamed</ansi-color>, <ansi-color #90>0 deleted</ansi-color>\)
%s
Worktrees:    0 total, 0 dirty
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo), remote)
	assert.Output(t, resp.Stdout, block)
}
```
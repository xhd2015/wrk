## Expected

- Exit code 0.
- `Remote:` value `diverged(2 commits)` is wrapped in red ANSI.
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
	if remote != "Remote:       diverged(2 commits)" {
		t.Fatalf("Remote: want diverged(2 commits), got %q", remote)
	}

	block := fmt.Sprintf(`---
version: 3
---
%s
%s
%s
Status:       clean
Remote:       <ansi-color red>diverged(2 commits)</ansi-color>
Worktrees:    0 total, 0 dirty
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo))
	assert.Output(t, resp.Stdout, block)
}
```
## Expected

- Exit code 0.
- `Remote:` value `needs push(+1 commit)` is wrapped in orange (`#33`) ANSI.
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

	block := fmt.Sprintf(`---
version: 2
---
%s
%s
%s
Status:       clean
Remote:       <ansi-color #33>needs push(+1 commit)</ansi-color>
Worktrees:    0 total, 0 dirty
`, colorProjectDirLine(t, req.MainRepo), colorStatusBranchLine(t, req.MainRepo), colorStatusCommitLine(t, req.MainRepo))
	assert.Output(t, resp.Stdout, block)
}
```
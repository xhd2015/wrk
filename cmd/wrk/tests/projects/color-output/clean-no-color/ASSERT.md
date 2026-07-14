## Expected

- Exit code 0.
- `Status:`, `Remote:`, and `Worktrees:` values have no red or orange ANSI (plain text).
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"strings"

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

	plain := stripANSI(resp.Stdout)
	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("stripped output still has escapes: %q", plain)
	}

	remote := colorCompareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	block := colorProjectStatusBlockTemplate(t, req.MainRepo, "clean", remote, "0 total, 0 dirty")
	assert.Output(t, plain, block)
}
```
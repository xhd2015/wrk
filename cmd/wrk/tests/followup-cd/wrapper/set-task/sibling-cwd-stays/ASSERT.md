## Expected

- Exit code 0.
- Stdout is the new worktree path.
- Old operated path gone; new path exists; sibling A still exists.
- Stderr has no `cd …` follow-up line.
- FinalPWD remains the sibling shell cwd (A).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify("new task"), 0)
	assert.Output(t, resp.Stdout, "---\nversion: 2\n---\n"+wantPath+"\n")
	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertFileExists(t, req.StartDir)
	if strings.Contains(resp.Stderr, "cd ") {
		t.Fatalf("expected no follow-up cd on stderr when sibling cwd survives; stderr=%q", resp.Stderr)
	}
	assertPathsEqual(t, resp.FinalPWD, req.StartDir)
}
```

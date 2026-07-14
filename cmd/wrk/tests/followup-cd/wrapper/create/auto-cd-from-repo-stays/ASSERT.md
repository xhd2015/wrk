## Expected

- Exit code 0.
- Create succeeds (stdout worktree path).
- Stderr has no `cd …` follow-up line (home gate closed for non-home shell cwd).
- FinalPWD remains the main-repo start directory.

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
	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assert.Output(t, resp.Stdout, "---\nversion: 2\n---\n"+wantPath+"\n")
	assertFileExists(t, wantPath)
	for _, line := range strings.Split(resp.Stderr, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "cd ") {
			t.Fatalf("non-home shell cwd must not print follow-up cd; got stderr line %q", line)
		}
	}
	assertPathsEqual(t, resp.FinalPWD, req.StartDir)
	assertPathsEqual(t, resp.FinalPWD, req.MainRepo)
}
```

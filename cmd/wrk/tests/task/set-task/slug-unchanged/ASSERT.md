## Expected

- Exit code 0.
- Stdout contains "task unchanged".
- Worktree directory still exists at old path.
- Branch still exists at old name.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(strings.TrimSpace(resp.Stdout), "task unchanged") {
		t.Fatalf("expected stdout to contain 'task unchanged', got %q", resp.Stdout)
	}

	// Worktree still exists at old path
	assertFileExists(t, req.WtDir)

	// Branch still exists
	branch := gitOutputIsolated(t, req.MainRepo, "rev-parse", "--abbrev-ref", "refs/heads/main-"+wrkDate+"-my-task")
	if branch == "" {
		t.Fatal("old branch should still exist")
	}
}
```

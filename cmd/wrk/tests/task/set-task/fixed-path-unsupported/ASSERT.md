## Expected

- Non-zero exit.
- Stdout empty (or no successful path print).
- Stderr explains that the directory/worktree is not wrk-managed / cannot parse directory name / unsupported for fixed paths.
- Worktree path and branch remain unchanged at the fixed path.

## Exit Code

- non-zero

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit for --set-task on fixed/non-wrk directory name")
	}

	combined := strings.ToLower(resp.Stdout + resp.Stderr)
	ok := strings.Contains(combined, "cannot parse") ||
		strings.Contains(combined, "directory") ||
		strings.Contains(combined, "unsupported") ||
		strings.Contains(combined, "not a wrk") ||
		strings.Contains(combined, "wrk worktree") ||
		strings.Contains(combined, "pattern")
	if !ok {
		t.Fatalf("expected error about non-wrk/fixed path/directory parse, got stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}

	// Fixed path still present; no rename performed.
	assertFileExists(t, req.WtDir)
	if filepath.Base(req.WtDir) != "wt" {
		t.Fatalf("expected fixed path basename wt, got %q", req.WtDir)
	}
	branch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, branch)
	assertBranchCheckedOutInWorktree(t, req.WtDir, branch)
}
```

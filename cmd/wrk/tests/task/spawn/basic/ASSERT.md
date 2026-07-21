```go
import (
	"path/filepath"
	"testing"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	slug := slugify("fix login bug")

	// stdout is the worktree path from Run
	wtDir := strings.TrimSpace(resp.Stdout)
	req.WtDir = wtDir
	assertFileExists(t, wtDir)
	assertGitFileIsWorktreeLink(t, wtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertBranchCheckedOutInWorktree(t, wtDir, req.WtBranch)
	assertContains(t, filepath.Base(wtDir), "-"+slug)
	assertContains(t, req.WtBranch, "-"+slug)

	basename := filepath.Base(wtDir)
	assertContains(t, basename, wrkDate+"-"+slug)
	assertContains(t, req.WtBranch, wrkDate+"-"+slug)
}
```

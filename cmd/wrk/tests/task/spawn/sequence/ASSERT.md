
```go
import (

	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	slug := slugify("same task")

	// First worktree (from Setup): no suffix
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertContains(t, filepath.Base(req.WtDir), "-"+slug)
	assertNotContains(t, filepath.Base(req.WtDir), slug+"-")

	// Second worktree (from Run): suffix 1
	wtDir2 := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wtDir2)
	expectedSecondBranch := branchNameWithTask("main", wrkDate, slug, 1)
	assertBranchExists(t, req.MainRepo, expectedSecondBranch)
	expectedSecond := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 1)
	if wtDir2 != expectedSecond {
		t.Fatalf("expected second worktree at %q, got %q", expectedSecond, wtDir2)
	}
}
```

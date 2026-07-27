
```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	slug := slugify("fix login")
	wtDir := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wtDir)
	assertGitFileIsWorktreeLink(t, wtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertBranchCheckedOutInWorktree(t, wtDir, req.WtBranch)

	expectedDir := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 1)
	if wtDir != expectedDir {
		t.Fatalf("expected dir %q, got %q", expectedDir, wtDir)
	}
}
```

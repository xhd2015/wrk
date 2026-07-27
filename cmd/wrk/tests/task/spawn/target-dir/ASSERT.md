
```go
import (
	"path/filepath"
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
	targetDir := filepath.Join(req.WorkRoot, "my-custom-dir")
	wtDir := strings.TrimSpace(resp.Stdout)
	if wtDir != targetDir {
		t.Fatalf("expected dir %q, got %q", targetDir, wtDir)
	}
	assertFileExists(t, wtDir)
	assertGitFileIsWorktreeLink(t, wtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertBranchCheckedOutInWorktree(t, wtDir, req.WtBranch)
	slug := slugify("fix it")
	assertContains(t, req.WtBranch, "-"+slug)
}
```

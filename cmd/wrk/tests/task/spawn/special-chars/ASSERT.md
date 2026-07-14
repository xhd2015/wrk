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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	expectedSlug := slugify(req.TaskDesc)
	if expectedSlug == "" {
		t.Fatal("slug should not be empty")
	}
	if expectedSlug != "fix-login-signup-urgent" {
		t.Fatalf("expected slug %q, got %q", "fix-login-signup-urgent", expectedSlug)
	}
	wtDir := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wtDir)
	assertGitFileIsWorktreeLink(t, wtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertContains(t, filepath.Base(wtDir), "-"+expectedSlug)
	assertContains(t, req.WtBranch, "-"+expectedSlug)
}
```

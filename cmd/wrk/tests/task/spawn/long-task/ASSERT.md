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
	runes := []rune(expectedSlug)
	if len(runes) > 64 {
		t.Fatalf("slug should be truncated to 64 runes, got %d: %q", len(runes), expectedSlug)
	}
	if len(runes) < 20 {
		t.Fatalf("slug too short (%d runes), truncation may have removed too much", len(runes))
	}
	wtDir := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wtDir)
	assertGitFileIsWorktreeLink(t, wtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertContains(t, filepath.Base(wtDir), "-"+expectedSlug)
}
```

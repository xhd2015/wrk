## Expected

- Exit 0; worktree created; path basename ≤ 255 bytes (budget fit applied).
- Outer agent-run invoked with `--dir` = worktree.
- Last argv element is exactly `/brainstorm ` + **full** original TaskDesc (not the fitted slug alone).
- Soft-cap slug alone would overflow path base for this basename (proves trim was needed).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	basename := strings.Repeat("r", 180)
	wt := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	base := filepath.Base(wt)
	if len(base) > 255 {
		t.Fatalf("path basename len %d > 255: %q", len(base), base)
	}
	fullSlug := slugify(req.TaskDesc)
	if utf8.RuneCountInString(fullSlug) < 20 {
		t.Fatalf("expected long soft-cap slug, got %q", fullSlug)
	}
	unfitted := basename + "-main-" + wrkDate + "-" + fullSlug
	if len(unfitted) <= 255 {
		t.Fatalf("fixture should need budget fit; unfitted len=%d", len(unfitted))
	}
	if strings.HasSuffix(base, "-"+fullSlug) {
		t.Fatalf("expected fitted (shortened) slug in path base; got %q fullSlug=%q", base, fullSlug)
	}

	// Agent must receive FULL taskDesc, not slug.
	args := assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertAgentArgvHasDir(t, args, wt)
	last := args[len(args)-1]
	wantPrompt := "/brainstorm " + req.TaskDesc
	if last != wantPrompt {
		t.Fatalf("agent prompt must be full taskDesc: want %q, got %q (full argv=%v)", wantPrompt, last, args)
	}
	// Explicitly reject slug-only prompt.
	if last == "/brainstorm "+fullSlug || last == "/brainstorm "+filepath.Base(wt) {
		t.Fatalf("agent must not receive slug-only prompt: %q", last)
	}
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
}
```

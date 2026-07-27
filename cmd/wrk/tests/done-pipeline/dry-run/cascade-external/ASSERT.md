## Expected

- Exit code 0.
- Stdout contains compact cascade dry-run line(s):
  `would: cascade merge-back <path>` (or equivalent containing both
  `would: cascade` and the external worktree path).
- No confirm prompt / non-TTY confirm errors.
- **External dep worktree still present** after dry-run (cascade must not mutate).
- Consumer worktree still present (primary dry-run).
- No real apply merge/remove of consumer or cascade targets.

## Side Effects

- None: external and consumer worktrees remain on disk.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)

	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	// Compact cascade plan vocabulary (locked product rule).
	out := resp.Stdout
	hasWouldCascade := strings.Contains(out, "would: cascade merge-back") ||
		strings.Contains(out, "would: cascade")
	if !hasWouldCascade {
		t.Fatalf("stdout missing cascade dry-run plan (want would: cascade merge-back <path>); got:\n%s", out)
	}
	// Path may be absolute, short, or relative — require a distinctive path fragment.
	extBase := filepath.Base(req.ExternalWtDir)
	if !strings.Contains(out, req.ExternalWtDir) && !strings.Contains(out, extBase) {
		t.Fatalf("cascade dry-run line should mention external path %q (or base %q); stdout:\n%s",
			req.ExternalWtDir, extBase, out)
	}

	// Zero cascade mutation: external still on disk and still a linked worktree of dep.
	assertFileExists(t, req.ExternalWtDir)
	assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	baseline := filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "external.sha")
	if _, err := os.Stat(baseline); err == nil {
		want := readBaselineSHA(t, req, "external.sha")
		got := revParseHEAD(t, req.ExternalWtDir)
		if got != want {
			t.Fatalf("external HEAD mutated under cascade dry-run: got %s want %s", got, want)
		}
	}

	// Consumer primary dry-run: still linked.
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)

	// Must not look like real cascade removal completed without plan-only mode.
	// (Real cascade removes external before parent continues.)
	assertFileExists(t, req.ExternalWtDir)
}
```

## Expected

- Exit code 0.
- Combined output contains phase headers **`==> cascade`** and **`==> own`**,
  with cascade appearing **before** own.
- Compact cascade plan under cascade phase:
  `would: cascade merge-back` (and external path or basename).
- Cascade plan line appears **after** `==> cascade` and **before** `==> own`
  (structure: items belong to cascade phase).
- No confirm prompt / non-TTY confirm errors.
- No ANSI escape sequences required for structure (assert stdout/stderr free of CSI).
- External dep worktree and consumer worktree still present (zero mutations).

## Side Effects

- None: external and consumer worktrees remain on disk.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoiseUX(t, resp)
	assertNoANSIEscape(t, resp.Stdout, "stdout")
	assertNoANSIEscape(t, resp.Stderr, "stderr")
	assertDonePhaseHeaders(t, resp)

	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	combined := resp.Stdout + "\n" + resp.Stderr
	// Compact cascade plan vocabulary (locked product rule).
	hasWouldCascade := strings.Contains(combined, "would: cascade merge-back") ||
		strings.Contains(combined, "would: cascade")
	if !hasWouldCascade {
		t.Fatalf("missing cascade dry-run plan (want would: cascade merge-back <path>); got:\n%s", combined)
	}
	extBase := filepath.Base(req.ExternalWtDir)
	if !strings.Contains(combined, req.ExternalWtDir) && !strings.Contains(combined, extBase) {
		t.Fatalf("cascade plan should mention external path %q (or base %q); combined:\n%s",
			req.ExternalWtDir, extBase, combined)
	}

	// Structure: cascade header → would: line → own header.
	idxCascade := strings.Index(combined, "==> cascade")
	idxOwn := strings.Index(combined, "==> own")
	idxWould := strings.Index(combined, "would: cascade")
	if idxWould < 0 {
		t.Fatalf("missing would: cascade in combined output:\n%s", combined)
	}
	if !(idxCascade < idxWould && idxWould < idxOwn) {
		t.Fatalf("want ==> cascade then would: cascade then ==> own; cascade@%d would@%d own@%d\n%s",
			idxCascade, idxWould, idxOwn, combined)
	}

	assertFileExists(t, req.ExternalWtDir)
	assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```

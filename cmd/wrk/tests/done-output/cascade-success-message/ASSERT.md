## Expected

- Exit code 0.
- Phase headers: `==> cascade` before `==> own`.
- Cascade section (between headers) includes a MergeBack **Message** for the external
  target — typically `worktree removed:` (contained remove-only). May also be
  `merged branch` if relation is ahead.
- Own section / stdout includes own MergeBack Message (`merged branch` or
  `worktree removed:`).
- External path or basename appears with cascade success output.
- External + consumer removed from disk.
- No ANSI required (pipe default).

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
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	assertNoANSIEscape(t, resp.Stdout, "stdout")
	assertNoANSIEscape(t, resp.Stderr, "stderr")
	assertDonePhaseHeaders(t, resp)

	combined := resp.Stdout + "\n" + resp.Stderr
	idxCascade := strings.Index(combined, "==> cascade")
	idxOwn := strings.Index(combined, "==> own")
	if idxCascade < 0 || idxOwn < 0 || idxCascade > idxOwn {
		t.Fatalf("phase order; cascade@%d own@%d\n%s", idxCascade, idxOwn, combined)
	}
	cascadeSection := combined[idxCascade:idxOwn]
	ownSection := combined[idxOwn:]

	// D5: cascade success must print Message (silent cascade today → RED).
	if !strings.Contains(cascadeSection, "worktree removed:") && !strings.Contains(cascadeSection, "merged branch") {
		t.Fatalf("cascade phase must print MergeBack Message (worktree removed: / merged branch); cascade section:\n%s\nfull stdout:\n%s",
			cascadeSection, resp.Stdout)
	}
	extBase := filepath.Base(req.ExternalWtDir)
	if !strings.Contains(cascadeSection, req.ExternalWtDir) && !strings.Contains(cascadeSection, extBase) {
		t.Fatalf("cascade Message should mention external path %q or base %q; cascade section:\n%s",
			req.ExternalWtDir, extBase, cascadeSection)
	}
	if !strings.Contains(ownSection, "worktree removed:") && !strings.Contains(ownSection, "merged branch") {
		t.Fatalf("own phase must print MergeBack Message; own section:\n%s", ownSection)
	}

	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```

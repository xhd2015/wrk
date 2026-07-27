## Expected

- Exit code 0.
- Phase headers **`==> cascade`** and **`==> own`** are **both absent** when there are
  **zero** nested cascade targets (including optional `==> cascade (0)` forms).
- No `would: cascade` plan lines (nothing to cascade).
- Primary MergeBack dry-run plan still present (`merge --ff-only`, worktree remove, branch -D).
- No confirm prompt noise; no ANSI required for structure.
- Zero mutations (wt still linked; main tip unchanged).

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoiseUX(t, resp)
	assertNoANSIEscape(t, resp.Stdout, "stdout")
	assertNoANSIEscape(t, resp.Stderr, "stderr")
	// Zero cascade targets: neither phase banner may appear.
	assertNoDonePhaseHeaders(t, resp)

	combined := resp.Stdout + "\n" + resp.Stderr
	// Zero targets: must not invent cascade merge-back plans.
	if strings.Contains(combined, "would: cascade merge-back") {
		t.Fatalf("zero-cascade dry-run must not print would: cascade merge-back; combined:\n%s", combined)
	}

	assertPrimaryDoneDryRunPlanned(t, resp.Stdout, req.WtBranch)
	// Zero mutations: wt still linked; feature-work not on main.
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
}
```

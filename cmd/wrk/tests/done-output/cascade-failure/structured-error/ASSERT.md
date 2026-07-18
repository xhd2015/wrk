## Expected

- Non-zero exit (cascade MergeBack failure).
- Combined output includes phase header **`==> cascade`** (cascade targets ≥ 1
  before fail). Do **not** require **`==> own`** — fail may stop mid-cascade.
- **Stderr** (primary hard-error stream) includes **`Error:`** prefix convention
  (substring; multi-line body OK).
- Error body includes the **external worktree path** or a distinctive path
  fragment (basename of `ExternalWtDir`).
- Framing is **not** bare `rebase conflict:` alone as the leading error line:
  after trim, stderr must not start with `rebase conflict:` without a prior
  `Error:` (git detail may follow indented or after structured fields).
- Prefer structure without ANSI in the error body.
- Consumer linked worktree still exists (own phase not applied after cascade fail).

## Side Effects

- Cascade did not complete remove of external (or may leave partial git state);
  consumer wt must remain.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on cascade merge-back failure; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	assertNoANSIEscape(t, resp.Stderr, "stderr")

	// Non-empty cascade: phase header for cascade is required; ==> own is optional
	// (own phase may never run after cascade failure).
	combined := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(combined, "==> cascade") {
		t.Fatalf("cascade-failure (targets ≥ 1) must print %q; stdout:\n%s\nstderr:\n%s",
			"==> cascade", resp.Stdout, resp.Stderr)
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("cascade failure stderr must include structured %q prefix; stderr:\n%s\nstdout:\n%s",
			"Error:", stderr, resp.Stdout)
	}

	extBase := filepath.Base(req.ExternalWtDir)
	if !strings.Contains(stderr, req.ExternalWtDir) && !strings.Contains(stderr, extBase) {
		// Path context may also appear on stdout phase lines; require path in stderr error body.
		// Allow combined only if Error: block is multi-stream — still prefer stderr.
		pathCombined := stderr + "\n" + resp.Stdout
		if !strings.Contains(pathCombined, req.ExternalWtDir) && !strings.Contains(pathCombined, extBase) {
			t.Fatalf("structured error must mention external path %q or base %q; stderr:\n%s\nstdout:\n%s",
				req.ExternalWtDir, extBase, stderr, resp.Stdout)
		}
	}

	// Must not lead with bare "rebase conflict:" as the sole framing.
	trimmed := strings.TrimSpace(stderr)
	if strings.HasPrefix(trimmed, "rebase conflict:") {
		t.Fatalf("stderr must not lead with bare %q; want Error: framing with path context; stderr:\n%s",
			"rebase conflict:", stderr)
	}
	// If rebase conflict detail is present, Error: must still appear first.
	if idxRC := strings.Index(stderr, "rebase conflict:"); idxRC >= 0 {
		idxErr := strings.Index(stderr, "Error:")
		if idxErr < 0 || idxErr > idxRC {
			t.Fatalf("when rebase conflict detail is present, Error: must appear before it; stderr:\n%s", stderr)
		}
	}

	// Own merge-back must not have completed after cascade failure.
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
}
```

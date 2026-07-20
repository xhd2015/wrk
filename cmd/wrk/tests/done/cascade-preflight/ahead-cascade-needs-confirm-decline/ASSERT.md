## Expected

- Cascade does **not** remove external (D3: default auto-yes must not silently confirm).
- Exit non-zero **or** exit 0 with explicit abort (`merge-back aborted`) — either is OK if
  external stays; prefer non-zero / `Error:` / confirmation language when no apply.
- Dep main must **not** contain the ahead fix commit (cascade merge did not apply).
- Consumer worktree still present.

Today bare/default path auto-yeses cascade → external gone → **RED**.

## Exit Code

- Prefer non-zero; exit 0 only if clearly aborted without mutation

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	// Mutation pin first: external must remain.
	assertFileExists(t, req.ExternalWtDir)
	assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)

	// Ahead fix must not have been merged into dep main.
	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	if strings.Contains(depLog, "dep fix on external worktree") {
		t.Fatalf("cascade must not merge ahead dep after decline/abort; dep log:\n%s\nstdout=%q stderr=%q",
			depLog, resp.Stdout, resp.Stderr)
	}

	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	aborted := strings.Contains(combined, "merge-back aborted") ||
		strings.Contains(combined, "aborted")
	if resp.ExitCode == 0 && !aborted {
		t.Fatalf("decline/confirm-required path must not succeed with mutations; exit=0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	// When non-zero, prefer structured Error: or confirm/prompt language (soft pin).
	if resp.ExitCode != 0 {
		if !strings.Contains(resp.Stderr, "Error:") &&
			!strings.Contains(combined, "confirm") &&
			!strings.Contains(combined, "proceed") &&
			!strings.Contains(combined, "terminal") &&
			!strings.Contains(combined, "stdin") {
			// Still OK if external preserved; soft diagnostic only when framing is empty.
			if strings.TrimSpace(resp.Stderr) == "" && strings.TrimSpace(resp.Stdout) == "" {
				t.Fatalf("expected Error:/confirm framing on non-zero abort; empty output")
			}
		}
	}

	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```

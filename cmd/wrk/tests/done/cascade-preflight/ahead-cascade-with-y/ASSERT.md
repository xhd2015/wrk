## Expected

- Exit code 0.
- Dep fix merged into dep main (`dep fix on external worktree` in log).
- External dependency worktree removed.
- Consumer linked worktree removed; branch deleted.
- Phase headers when cascade targets ≥ 1: `==> cascade` before `==> own`.
- On real success, cascade MergeBack **Message** should appear on stdout (D5) —
  e.g. `merged branch` for ahead cascade and/or `worktree removed:` for own.
  Message pin is soft-failing today (silent cascade) → may contribute **RED** until D5.

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 with -y cascade+own; got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on external worktree")

	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)

	combined := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(combined, "==> cascade") {
		t.Fatalf("with cascade targets must print %q; combined:\n%s", "==> cascade", combined)
	}
	if !strings.Contains(combined, "==> own") {
		t.Fatalf("with cascade targets must print %q; combined:\n%s", "==> own", combined)
	}
	idxC := strings.Index(combined, "==> cascade")
	idxO := strings.Index(combined, "==> own")
	if idxC > idxO {
		t.Fatalf("want ==> cascade before ==> own; cascade@%d own@%d\n%s", idxC, idxO, combined)
	}

	// D5: each cascade success prints MergeBack Message (ahead → merged branch).
	// Own also prints a Message. Require at least one success message token on stdout.
	if !strings.Contains(resp.Stdout, "merged branch") && !strings.Contains(resp.Stdout, "worktree removed:") {
		t.Fatalf("real success must print MergeBack Message(s) on stdout (merged branch / worktree removed:); stdout:\n%s",
			resp.Stdout)
	}
	// Stronger D5 pin: cascade phase should emit a message before ==> own (not silent cascade).
	// Between ==> cascade and ==> own expect a Message line (merged branch or worktree removed:).
	cascadeSection := combined
	if idxC >= 0 && idxO > idxC {
		cascadeSection = combined[idxC:idxO]
	}
	if !strings.Contains(cascadeSection, "merged branch") && !strings.Contains(cascadeSection, "worktree removed:") {
		t.Fatalf("D5: cascade phase must print MergeBack Message before ==> own; cascade section:\n%s\nfull:\n%s",
			cascadeSection, combined)
	}
}
```

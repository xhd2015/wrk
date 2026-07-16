## Expected

- Exit code 0 (full dry-run plan; no confirm prompts).
- Stderr must **not** contain `mutually exclusive`.
- Gen-commit dry plan is visible:
  - stdout includes mock B style `would generate commit message` for staged file(s), **and/or**
  - stderr includes `would:` + `git commit` plan line.
- Must not execute a real commit (`Running git commit...` absent; HEAD subject unchanged).
- Primary MergeBack dry-run planned commands present (`merge --ff-only <WtBranch>`, `worktree remove`, `branch -D`).
- Side effects: **zero mutations** — wt still linked; main HEAD unchanged; no real commit; staged file still staged optional but HEAD subject must match baseline.

## Side Effects

- Plan only (no commit, no merge-back apply).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("gen-commit pre + done dry-run still mutually exclusive; stderr=%q", resp.Stderr)
	}
	assertNoConfirmPromptNoise(t, resp)

	// Gen-commit dry plan (library mock B and/or would: git commit).
	outAll := resp.Stdout + "\n" + resp.Stderr
	hasMockB := strings.Contains(outAll, "would generate commit message")
	hasWouldCommit := strings.Contains(strings.ToLower(resp.Stderr), "would:") &&
		strings.Contains(resp.Stderr, "git commit")
	if !hasMockB && !hasWouldCommit {
		t.Fatalf("expected gen-commit dry-run plan (mock B and/or would: git commit); stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}

	// Primary merge-back dry plan.
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)
	assertDoneDryRunZeroMutations(t, req)

	// HEAD subject on worktree unchanged (no real commit).
	wantSubject := strings.TrimSpace(readBaselineSHA(t, req, "wt.head-subject"))
	// readBaselineSHA works for any single-line baseline file under the dry-run baseline dir.
	gotSubject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	if gotSubject != wantSubject {
		t.Fatalf("worktree HEAD subject changed under dry-run gen-commit: before=%q after=%q",
			wantSubject, gotSubject)
	}

	// Staged file should still be staged (dry-run does not commit).
	cached := gitOutputIsolated(t, req.WtDir, "diff", "--cached", "--name-only")
	if !strings.Contains(cached, "staged-for-commit.go") {
		t.Fatalf("staged-for-commit.go should remain staged under dry-run; cached=%q", cached)
	}
}
```

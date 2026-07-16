## Expected

- Exit code 0.
- No confirm prompt / mutual exclusion.
- **Gen-commit pre** dry plan visible (mock B and/or `would:` + `git commit`); no real commit.
- Primary MergeBack DryRun plan present.
- Post stages: sync → tag-next → push → **reinstall** (order markers).
- Full stage order: **gen-commit < primary < sync < tag < push < reinstall**.
- Reinstall dry vocabulary present; no real install (`reinstalled ` absent).
- Zero mutations: wt linked; tags/origin unchanged; GOBIN stub unchanged; HEAD subject + staged file preserved.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("flag layer still rejects full pre+posts+reinstall dry-run; stderr=%q", resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
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

	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)

	syncBlock := wouldSyncDistributeOne(req.Wt2Branch, 1)
	tagBlock := tagNextRootBumpPlanStdout()
	pushBlock := wouldPushMainOrigin("v0.0.2")
	for _, part := range []string{
		strings.TrimSpace(syncBlock),
		strings.TrimSpace(tagBlock),
		strings.TrimSpace(pushBlock),
	} {
		if !strings.Contains(resp.Stdout, part) {
			t.Fatalf("stdout missing post-stage block %q\nfull stdout:\n%s", part, resp.Stdout)
		}
	}
	assertReinstallDryRunPresent(t, resp.Stdout)

	// Stage order: gen-commit → primary → sync → tag → push → reinstall.
	// Gen-commit markers may land on stdout (mock B) and/or stderr (would: git commit).
	idxGen := -1
	if i := strings.Index(outAll, "would generate commit message"); i >= 0 {
		idxGen = i
	}
	if i := strings.Index(resp.Stderr, "git commit"); i >= 0 {
		// Prefer earliest gen marker across streams; stderr offset is independent of stdout.
		// For cross-stream order we only require gen-commit evidence + stdout stage order below.
		_ = i
		if idxGen < 0 {
			idxGen = 0 // gen-commit evidence exists on stderr; treat as "before" stdout stages
		}
	}
	if idxGen < 0 {
		t.Fatalf("missing gen-commit stage marker in combined output\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}

	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxSync := strings.Index(resp.Stdout, "would: synced:")
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxPush := strings.Index(resp.Stdout, "would: git push origin main")
	idxRe := strings.Index(resp.Stdout, "would: go install ./cmd/present")
	if idxMerge < 0 || idxSync < 0 || idxTag < 0 || idxPush < 0 || idxRe < 0 {
		t.Fatalf("missing stage markers merge=%d sync=%d tag=%d push=%d reinstall=%d\n%s",
			idxMerge, idxSync, idxTag, idxPush, idxRe, resp.Stdout)
	}
	if !(idxMerge < idxSync && idxSync < idxTag && idxTag < idxPush && idxPush < idxRe) {
		t.Fatalf("stage order want primary < sync < tag < push < reinstall; got merge=%d sync=%d tag=%d push=%d reinstall=%d\n%s",
			idxMerge, idxSync, idxTag, idxPush, idxRe, resp.Stdout)
	}
	// When mock B is on stdout, it must appear before the primary merge plan.
	if i := strings.Index(resp.Stdout, "would generate commit message"); i >= 0 && !(i < idxMerge) {
		t.Fatalf("gen-commit mock B must precede primary merge plan; gen=%d merge=%d\n%s",
			i, idxMerge, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "pushed main →")
	assertNotContains(t, resp.Stdout, "reinstalled ")
	if strings.Contains(resp.Stdout, "tagged v0.0.2") {
		t.Fatalf("dry-run must not print apply 'tagged' lines; stdout=%q", resp.Stdout)
	}

	assertDoneDryRunZeroMutations(t, req)
	assertStubPresentUnchanged(t, filepath.Join(req.WorkRoot, "gobin"))

	// HEAD subject on worktree unchanged (no real commit).
	wantSubject := strings.TrimSpace(readBaselineSHA(t, req, "wt.head-subject"))
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

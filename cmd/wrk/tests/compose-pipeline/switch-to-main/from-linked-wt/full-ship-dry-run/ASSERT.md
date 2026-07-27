
## Expected

- Exit 0; no mutual exclusion; no confirm noise.
- Gen-commit dry plan present (mock B and/or `would:` + `git commit`).
- Done MergeBack dry plan: `merge --ff-only <WtBranch>`, `worktree remove`, `branch -D`.
- Post stages after done (activeRoot would be main): sync, tag-next, push, reinstall.
- Fixed order markers: gen-commit evidence before merge; merge < sync < tag < push < reinstall.
- Zero mutations: wt still linked; no v0.0.2; origin unchanged; staged file remains; GOBIN stub unchanged.

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
	assertNoMutexReject(t, resp.Stderr)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertAPNoConfirmNoise(t, resp)

	outAll := resp.Stdout + "\n" + resp.Stderr
	hasMockB := strings.Contains(outAll, "would generate commit message")
	hasWouldCommit := strings.Contains(strings.ToLower(resp.Stderr), "would:") &&
		strings.Contains(resp.Stderr, "git commit")
	if !hasMockB && !hasWouldCommit {
		t.Fatalf("expected gen-commit dry-run plan; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stdout, "merge --ff-only "+req.WtBranch)
	assertContains(t, resp.Stdout, "worktree remove")
	assertContains(t, resp.Stdout, "branch -D "+req.WtBranch)

	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxSync := strings.Index(resp.Stdout, "would: synced:")
	if idxSync < 0 {
		idxSync = strings.Index(resp.Stdout, "would: "+req.Wt2Branch)
	}
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxPush := strings.Index(resp.Stdout, "would: git push origin main")
	idxRe := strings.Index(resp.Stdout, "would: go install ./cmd/present")
	if idxMerge < 0 || idxTag < 0 || idxPush < 0 || idxRe < 0 {
		t.Fatalf("missing stage markers merge=%d tag=%d push=%d reinstall=%d\n%s",
			idxMerge, idxTag, idxPush, idxRe, resp.Stdout)
	}
	// sync may be weak if only one wt behind; require tag after merge and reinstall last among posts
	if !(idxMerge < idxTag && idxTag < idxPush && idxPush < idxRe) {
		t.Fatalf("want merge < tag < push < reinstall; merge=%d tag=%d push=%d re=%d\n%s",
			idxMerge, idxTag, idxPush, idxRe, resp.Stdout)
	}
	if idxSync >= 0 && !(idxMerge < idxSync && idxSync < idxTag) {
		t.Fatalf("want merge < sync < tag; merge=%d sync=%d tag=%d\n%s",
			idxMerge, idxSync, idxTag, resp.Stdout)
	}
	if i := strings.Index(resp.Stdout, "would generate commit message"); i >= 0 && !(i < idxMerge) {
		t.Fatalf("gen-commit must precede done plan; gen=%d merge=%d\n%s", i, idxMerge, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "reinstalled ")
	assertAPDryRunZeroMutationsLinked(t, req)
	assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))

	wantSubject := readAPBaseline(t, req, "wt.head-subject")
	gotSubject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	if gotSubject != wantSubject {
		t.Fatalf("worktree HEAD subject changed under dry-run: before=%q after=%q", wantSubject, gotSubject)
	}
	cached := gitOutputIsolated(t, req.WtDir, "diff", "--cached", "--name-only")
	if !strings.Contains(cached, "staged-for-commit.go") {
		t.Fatalf("staged-for-commit.go should remain staged; cached=%q", cached)
	}
}
```

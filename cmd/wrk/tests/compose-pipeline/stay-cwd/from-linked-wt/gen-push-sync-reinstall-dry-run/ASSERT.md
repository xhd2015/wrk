
## Expected

- Exit 0; not mutually exclusive (multi-stage without primary allowed).
- Gen-commit dry plan present.
- Push plan targets the **worktree branch** (not only main), e.g. `would: git push origin <WtBranch>`.
- Reinstall plan present (`would: go install ./cmd/present`).
- Fixed relative order among requested stages: gen-commit before push/sync/reinstall; push/sync before reinstall is acceptable; no tag-next.
- Zero mutations: staged remains or subject preserved under dry-run; stub unchanged; no v0.0.2.

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

	outAll := resp.Stdout + "\n" + resp.Stderr
	hasMockB := strings.Contains(outAll, "would generate commit message")
	hasWouldCommit := strings.Contains(strings.ToLower(resp.Stderr), "would:") &&
		strings.Contains(resp.Stderr, "git commit")
	if !hasMockB && !hasWouldCommit {
		t.Fatalf("expected gen-commit dry plan; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	// Push should follow WT activeRoot → feature branch, not silently main-only.
	wantPush := "would: git push origin " + req.WtBranch
	if !strings.Contains(resp.Stdout, wantPush) &&
		!strings.Contains(resp.Stdout, "git push origin "+req.WtBranch) {
		// Accept alternate would: vocabulary if branch name appears with push.
		if !(strings.Contains(resp.Stdout, "would: git push") && strings.Contains(resp.Stdout, req.WtBranch)) {
			t.Fatalf("expected push plan for worktree branch %s; stdout=%q", req.WtBranch, resp.Stdout)
		}
	}

	assertReinstallDryRunAP(t, resp.Stdout)
	assertNotContains(t, resp.Stdout, "1 tag planned")
	assertNotContains(t, resp.Stdout, "merge --ff-only")
	assertNotContains(t, resp.Stdout, "reinstalled ")
	assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))

	wantSubject := readAPBaseline(t, req, "wt.head-subject")
	gotSubject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	if gotSubject != wantSubject {
		t.Fatalf("wt HEAD subject changed under dry-run: before=%q after=%q", wantSubject, gotSubject)
	}
}
```


## Expected

- Exit 0; no mutual exclusion; no confirm noise.
- Not the bare-`--main` nested-shell path (pipeline plans present).
- Tag-next plan on main (`1 tag planned`); push plan for main; reinstall plan.
- Fixed order among posts: tag before push before reinstall (sync may be weak/empty).
- No done/merge-back plan lines (`merge --ff-only`, `worktree remove`, `branch -D`).
- Zero mutations: linked wt still exists; no v0.0.2; origin unchanged; GOBIN stub unchanged.

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

	if !strings.Contains(resp.Stdout, "1 tag planned") {
		t.Fatalf("expected tag-next plan under --main from linked wt; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "would: git push origin main") {
		t.Fatalf("expected push plan for main; stdout=%q", resp.Stdout)
	}
	assertReinstallDryRunAP(t, resp.Stdout)

	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	idxPush := strings.Index(resp.Stdout, "would: git push origin main")
	idxRe := strings.Index(resp.Stdout, "would: go install ./cmd/present")
	if !(idxTag < idxPush && idxPush < idxRe) {
		t.Fatalf("want tag < push < reinstall; tag=%d push=%d re=%d\n%s",
			idxTag, idxPush, idxRe, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "merge --ff-only")
	assertNotContains(t, resp.Stdout, "worktree remove")
	assertNotContains(t, resp.Stdout, "branch -D")
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "reinstalled ")

	assertAPDryRunZeroMutationsLinked(t, req)
	assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))
}
```

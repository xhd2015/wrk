
## Expected

- Exit 0; not mutually exclusive.
- Stderr contains a notice that `--main` is not necessary / already at main, and that wrk is continuing (pipeline not aborted).
- Tag-next plan present (`1 tag planned`); push + reinstall plans; order tag < push < reinstall.
- No done/merge-back plan lines.
- Zero mutations (no v0.0.2; stub unchanged).

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

	se := resp.Stderr
	lower := strings.ToLower(se)
	// Prefer product wording: --main is not necessary (already at main …); continuing
	hasNotNecessary := strings.Contains(lower, "not necessary") ||
		strings.Contains(lower, "unnecessary") ||
		strings.Contains(lower, "not needed")
	hasAlreadyMain := strings.Contains(lower, "already") && strings.Contains(lower, "main")
	hasContinue := strings.Contains(lower, "continu") // continuing / continue
	if !(hasNotNecessary || hasAlreadyMain) {
		t.Fatalf("stderr should notice redundant --main / already at main; got %q", se)
	}
	if !hasContinue && !hasNotNecessary {
		t.Fatalf("stderr should indicate continuing after notice; got %q", se)
	}
	if !strings.Contains(se, "--main") && !strings.Contains(lower, "main") {
		t.Fatalf("stderr notice should mention main / --main; got %q", se)
	}

	if !strings.Contains(resp.Stdout, "1 tag planned") {
		t.Fatalf("expected tag-next plan after notice; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
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
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "reinstalled ")
	if tagRefExistsAP(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not be created under dry-run")
	}
	assertStubPresentAP(t, filepath.Join(req.WorkRoot, "gobin"))
}
```

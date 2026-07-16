## Expected

- Exit code 0.
- No confirm prompt / mutual exclusion.
- Primary MergeBack DryRun plan present.
- Post stages: sync → tag-next → push → **reinstall** (order markers).
- Reinstall dry vocabulary present; no real install (`reinstalled ` absent).
- Zero mutations: wt linked; tags/origin unchanged; GOBIN stub unchanged.

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
		t.Fatalf("flag layer still rejects full-combo+reinstall dry-run; stderr=%q", resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
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

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "pushed main →")
	assertNotContains(t, resp.Stdout, "reinstalled ")
	if strings.Contains(resp.Stdout, "tagged v0.0.2") {
		t.Fatalf("dry-run must not print apply 'tagged' lines; stdout=%q", resp.Stdout)
	}

	assertDoneDryRunZeroMutations(t, req)
	assertStubPresentUnchanged(t, filepath.Join(req.WorkRoot, "gobin"))
}
```

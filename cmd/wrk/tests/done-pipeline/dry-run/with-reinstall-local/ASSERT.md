## Expected Output (stage vocabulary)

Primary MergeBack DryRun planned commands, then reinstall dry plan:

```
  # main: fast forward
  git -C <main> merge --ff-only <WtBranch>
  …
would: go install ./cmd/present
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit code 0.
- No confirm prompt / non-TTY confirm errors.
- No `mutually exclusive` on stderr.
- Primary: MergeBack DryRun plan for ahead+remove.
- Reinstall stage present after primary (`would: go install ./cmd/present`, summary).
- No real post stages not requested (no sync/tag/push apply vocabulary).
- Zero mutations: wt still linked; main HEAD unchanged; GOBIN/present remains stub.

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
		t.Fatalf("flag layer still rejects done+reinstall dry-run; stderr=%q", resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoConfirmPromptNoise(t, resp)
	assertPrimaryMergeBackDryRunPlanned(t, resp.Stdout, req.WtBranch)
	assertReinstallDryRunPresent(t, resp.Stdout)

	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxRe := strings.Index(resp.Stdout, "would: go install ./cmd/present")
	if idxMerge < 0 || idxRe < 0 {
		t.Fatalf("missing stage markers merge=%d reinstall=%d\n%s", idxMerge, idxRe, resp.Stdout)
	}
	if !(idxMerge < idxRe) {
		t.Fatalf("stage order want primary < reinstall; merge=%d reinstall=%d\n%s",
			idxMerge, idxRe, resp.Stdout)
	}

	// Alone reinstall tail: no sync/tag/push apply or other post stages.
	assertNotContains(t, resp.Stdout, "would: synced:")
	assertNotContains(t, resp.Stdout, "tag planned")
	assertNotContains(t, resp.Stdout, "would: git push")
	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "reinstalled ")

	assertDoneDryRunZeroMutations(t, req)
	assertStubPresentUnchanged(t, filepath.Join(req.WorkRoot, "gobin"))
}
```

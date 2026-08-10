## Expected

- Exit code 0.
- Final consumer go.mod:
  - **no** external `replace` for `example.com/dot-pkgs`
  - `require example.com/dot-pkgs v0.0.1` (D3 keep-current)
- Git history on consumer main: cascade pin commit
  `wrk: cascade pin example.com/dot-pkgs @ v0.0.1` is an **ancestor** of the
  feature gen-commit (`feat: add feature`) — pin before feature (D7 / B1).
- Feature WIP content (`FEATURE_WIP.md`) landed via gen-commit (not in pin commit).
- Pin commit touches **only** go.mod/go.sum.
- Pre-commit hook never blocks (pin removed external replace first).

## Side Effects

- Cascade pin auto-commit may use `--no-verify` (tool-authored deps commit).
- Feature gen-commit must **not** need `--no-verify` for local-replace hooks.
- `--merge-back` keeps linked consumer worktree; pin+feature land on branch/main.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		// Today RED: gen-commit hits no-local-replace hook while replace remains,
		// or peels/commits before cascade pin reordering.
		t.Fatalf("T1 pin-before-gen-commit: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required")
	}

	// Require + drop external replace (D3 keep-current @ v0.0.1).
	assertConsumerRequireAndNoExternalReplace(t, req)

	// History: pin commit before feature gen-commit (D7).
	hist := historyRepoForConsumer(t, req)
	assertCascadePinBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)

	// go.mod/go.sum should not leave uncommitted pin dirt after successful apply
	// with --add-all (feature+pin both committed).
	assertGoModCommittedClean(t, hist)

	// Sanity: pin commit message style.
	log := gitOutputIsolated(t, hist, "log", "--oneline", "-15")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

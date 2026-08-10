## Expected

- Exit code 0.
- Final monorepo go.mod:
  - **no** external `replace` for `example.com/dot-pkgs`
  - `require example.com/dot-pkgs v0.0.1` (D3 keep-current / ready free)
  - **keep** intra `replace … => ./pkgs/shared` (not force-dropped)
- Git history on monorepo main: cascade pin commit
  `wrk: cascade pin example.com/dot-pkgs @ v0.0.1` is an **ancestor** of the
  feature gen-commit (`feat: add feature`) — ready external pin before feature
  on freeHost peel (D7 / T-M1).
- Feature WIP content (`FEATURE_WIP.md`) landed via gen-commit (not in pin commit).
- External pin commit touches **only** go.mod/go.sum (no scoop of staged feature files).
- Pre-commit hook never blocks (external replace removed before feature commit).

## Side Effects

- Cascade pin auto-commit may use `--no-verify` (tool-authored deps commit).
- Feature gen-commit must **not** need `--no-verify` for local-replace hooks.
- Monorepo freeHost may still land/tag/pin intra free modules after prelude;
  free-first monorepo tag order after land is out of scope for this leaf.
- `--merge-back` keeps linked monorepo worktree; pin+feature land on branch/main.

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
		// Today RED: freeHost early peel gen-commits with external replace still
		// present → no-local-replace hook fails (B1 freeHost blocks pure deferral).
		t.Fatalf("T-M1 monorepo freeHost pin-before-gen-commit: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required")
	}

	// Require + drop external replace (D3 keep-current @ v0.0.1).
	assertConsumerRequireAndNoExternalReplace(t, req)

	// Intra keep-local must not be force-dropped with the external pin.
	assertIntraSharedReplaceKept(t, req)

	// History: ready external pin commit before feature gen-commit (D7 / T-M1).
	// Use dep-scoped pin SHA (monorepo may also pin shared later).
	hist := historyRepoForConsumer(t, req)
	assertCascadePinForDepBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	// Pin must not scoop pre-staged FEATURE_WIP (commit --only go.mod/go.sum).
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)

	// go.mod/go.sum should not leave uncommitted pin dirt after successful apply
	// with --add-all (feature+pin both committed).
	assertGoModCommittedClean(t, hist)

	// Sanity: pin subject for external dep + feature subject in log.
	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix+req.LeafModulePath) &&
		!strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

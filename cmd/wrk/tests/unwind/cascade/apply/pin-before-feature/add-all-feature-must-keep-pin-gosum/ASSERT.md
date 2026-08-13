## Expected

- Exit code 0.
- Consumer require `example.com/dot-pkgs v0.0.1`; no droppable external replace.
- Pin commit is an ancestor of feature gen-commit (`feat: add feature`).
- Pin commit `go.sum` contains `example.com/dot-pkgs v0.0.1` hash lines.
- Feature commit **tree** `go.sum` still contains those hash lines (must not
  scoop the hashless replace-tidy restore).
- HEAD `go.sum` still contains those hash lines.
- `FEATURE_WIP.md` landed; pin commit only `go.mod`/`go.sum`.

## Side Effects

- Partial-edit save/restore of WIP `go.sum` still runs; the feature commit
  must not publish that restored blob.
- `--merge-back` keeps the linked consumer worktree.

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
		t.Fatalf("GS1 add-all must keep pin go.sum hashes: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.MainRepo == "" || req.LeafModulePath == "" || req.ExpectedPinVersion == "" {
		t.Fatal("MainRepo, LeafModulePath, and ExpectedPinVersion required")
	}

	assertConsumerRequireAndNoExternalReplace(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)

	assertPinCommitGoSumHasDep(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertFeatureTreeGoSumHasDep(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertHEADGoSumHasDep(t, hist, req.LeafModulePath, req.ExpectedPinVersion)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-15")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("GS1: expected cascade pin commit in log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("GS1: expected feature subject %q in log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

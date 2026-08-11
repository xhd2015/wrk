## Expected

- Exit code 0.
- **Free skip contract** (primary RED lock):
  - Free main HEAD unchanged from baseline (still at LatestTag `v0.0.1`).
  - Tag `v0.0.1` exists; tag `v0.0.2` (next) must **not** exist on free.
  - No cascade/apply `tag-next` line for free module `example.com/dot-pkgs`.
  - Free external linked WT not peeled (`would: peel` / apply banner absent).
  - Free has no new product commit after baseline (HEAD == `v0.0.1` object).
- Consumer (same success path as T1 when pin-before-feature works):
  - require free @ `v0.0.1`; no droppable external replace
  - pin auto-commit then feature gen-commit when exit 0

## Side Effects

- Consumer may land/merge-back; free must not be mutated for release.
- Pin may use `--no-verify`; free gen-commit must not run.

## Errors

- None on success path.
- **RED today** if free gets next tag, free peel, or free HEAD moves.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("A-clean-tag: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	// --- Free skip: no next tag, no free tag-next, HEAD frozen at LatestTag ---
	freeMain := req.SecondRepo
	if tagRefExists(t, freeMain, unwindApplyNextTag) {
		t.Fatalf("A-clean-tag: free must not create next tag %s when already at %s\nfree tags:\n%s\ncombined:\n%s",
			unwindApplyNextTag, unwindApplyOldTag,
			gitOutputIsolated(t, freeMain, "tag", "--list"), out)
	}
	if !tagRefExists(t, freeMain, unwindApplyOldTag) {
		t.Fatalf("A-clean-tag: free LatestTag %s missing after apply", unwindApplyOldTag)
	}
	assertNoCascadeTagNextForModule(t, out, req.LeafModulePath)

	headAfter := revParseRef(t, freeMain, "HEAD")
	tagObj := revParseRef(t, freeMain, unwindApplyOldTag)
	if headAfter != tagObj {
		t.Fatalf("A-clean-tag: free HEAD must stay at LatestTag %s\nHEAD=%s\ntag=%s\nlog:\n%s",
			unwindApplyOldTag, headAfter, tagObj,
			gitOutputIsolated(t, freeMain, "log", "--oneline", "-10"))
	}
	baselinePath := filepath.Join(req.WorkRoot, "_free_head.sha")
	if data, err := os.ReadFile(baselinePath); err == nil {
		baseline := strings.TrimSpace(string(data))
		if baseline != "" && headAfter != baseline {
			t.Fatalf("A-clean-tag: free HEAD moved (baseline=%s after=%s)\nlog:\n%s",
				baseline, headAfter,
				gitOutputIsolated(t, freeMain, "log", "--oneline", "-10"))
		}
	}

	// Free external linked WT must not peel (clean free out of PeelOrder).
	if req.DepsLinkedWtDir != "" {
		skipped := peelDisplay(t, req, req.DepsLinkedWtDir)
		if hasPeelLine(out, skipped) {
			t.Fatalf("A-clean-tag: clean free must not peel %q\nstdout:\n%s",
				peelLine(skipped), resp.Stdout)
		}
		// Free linked checkout must remain clean porcelain (no free gen-commit dirt).
		st := strings.TrimSpace(gitOutputIsolated(t, req.DepsLinkedWtDir, "status", "--porcelain"))
		if st != "" {
			t.Fatalf("A-clean-tag: free external WT must stay clean; porcelain:\n%s", st)
		}
	}

	// Free must not land a feature-style gen-commit on free history.
	freeLog := gitOutputIsolated(t, freeMain, "log", "--oneline", "-15")
	if strings.Contains(freeLog, cascadeFeatureCommitSubject) {
		t.Fatalf("A-clean-tag: free must not receive consumer feature gen-commit\nlog:\n%s", freeLog)
	}

	// --- Consumer path (same as T1 when pin-before-feature is healthy) ---
	assertNoLocalReplaceGenCommitFail(t, out)
	assertConsumerRequireAndNoExternalReplace(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-15")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in consumer log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

## Expected

- Exit code 0.
- Free-first ship:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Order:** free `tag-next … @ v0.0.2` before consumer pin of free @ next.
- Consumer after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace (including **absolute** NewPath)
  - cascade pin auto-commit present; pin before feature gen-commit
  - `FEATURE_WIP.md` landed; pin commit only go.mod/go.sum
- No no-local-replace hook failure (abs path treated as forbidden while present).
- No `unknown revision`.

## Side Effects

- Absolute-path replace is droppable cross-repo; pin owns the drop (D1).
- Feature gen-commit runs only after pin removed abs replace (hook OK).
- `--merge-back` keeps linked consumer; `--push` publishes free when clear.

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
		t.Fatalf("D1 absolute-path external replace: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	assertNoLocalReplaceGenCommitFail(t, out)
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("D1: unknown revision after apply\ncombined:\n%s", out)
	}

	assertLeafMainAdvancedAndTagged(t, req)
	assertFreeTagNextBeforeConsumerPinOfFree(t, out)
	assertConsumerRequireAndNoExternalReplace(t, req)
	assertNoAbsoluteExternalReplaceForLeaf(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)

	// Feature tree must not reintroduce absolute replace for free.
	if featSHA := featureCommitSHA(t, hist, cascadeFeatureCommitSubject); featSHA != "" {
		featGoMod := gitOutputIsolated(t, hist, "show", featSHA+":go.mod")
		for _, line := range strings.Split(featGoMod, "\n") {
			trim := strings.TrimSpace(line)
			if !strings.Contains(trim, req.LeafModulePath) || !strings.Contains(trim, "=>") {
				continue
			}
			parts := strings.SplitN(trim, "=>", 2)
			if len(parts) != 2 {
				continue
			}
			newPath := strings.TrimSpace(parts[1])
			if strings.HasPrefix(newPath, "/") || strings.Contains(newPath, "./external/") || strings.HasPrefix(newPath, "../") {
				t.Fatalf("D1: feature gen-commit must not carry external/abs replace for %s\n%s",
					req.LeafModulePath, featGoMod)
			}
		}
	}

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in consumer log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

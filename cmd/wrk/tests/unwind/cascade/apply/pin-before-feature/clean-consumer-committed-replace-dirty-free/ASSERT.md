## Expected

- Exit code 0.
- Free-first ship:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Order:** free `tag-next … @ v0.0.2` before consumer pin of free @ next.
- Consumer after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace
  - cascade pin auto-commit present (`wrk: cascade pin …` for free @ next)
  - pin commit touches only go.mod/go.sum
- No consumer feature gen-commit required (no FEATURE_WIP seeded).
- No no-local-replace hook failure; no `unknown revision`.
- go.mod/go.sum clean on consumer history checkout.

## Side Effects

- Peel free only (consumer clean → not in dirty peel order).
- Cascade pin owns the replace drop on consumer Path/main.
- `--merge-back` keeps linked consumer WT; `--push` publishes free when clear.

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
		t.Fatalf("A4 clean consumer committed replace: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	assertNoLocalReplaceGenCommitFail(t, out)
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("A4: unknown revision after apply\ncombined:\n%s", out)
	}

	assertLeafMainAdvancedAndTagged(t, req)
	assertFreeTagNextBeforeConsumerPinOfFree(t, out)
	assertConsumerRequireAndNoExternalReplace(t, req)

	// A4: consumer is clean porcelain → not peeled/landed. Cascade pin commits on
	// the linked Path (WtDir), not necessarily main. Prefer WtDir for history.
	hist := req.WtDir
	if hist == "" {
		hist = historyRepoForConsumer(t, req)
	}
	assertCascadePinCommitPresent(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertGoModCommittedClean(t, hist)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	// A4: no FEATURE_WIP — if a feat: commit appears, it must not carry external replace.
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
			if strings.Contains(newPath, "./external/") || strings.HasPrefix(newPath, "../") ||
				strings.HasPrefix(newPath, "/") {
				t.Fatalf("A4: feature gen-commit must not carry external replace for %s\n%s",
					req.LeafModulePath, featGoMod)
			}
		}
	}
}
```

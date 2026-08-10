## Expected

- Exit code 0.
- Free-first ship for dirty free leaf:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Order (T-spl):** combined stdout/stderr shows
  `tag-next example.com/dot-pkgs @ v0.0.2` **before**
  consumer pin of free @ `v0.0.2` (cascade pin after free tag; monorepo not
  early-peeled as false freeHost).
- Consumer monorepo after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace for that module
  - cascade pin auto-commit present (`wrk: cascade pin …` for free @ next)
  - keep intra replace `=> ./pkgs/shared` (not force-dropped)
- Replace-only consumer: **must not** fail no-local-replace pre-commit; **must
  not** land a feature/gen-commit whose purpose is committing the local
  worktree replace (pin auto-commit owns the drop).
- go.mod/go.sum clean on consumer history checkout after land.
- Combined output must **not** contain `unknown revision`.

## Side Effects

- Early peel: free only (true tag host). Monorepo pure pin-consumer deferred.
- Cascade may also pin shared @ LatestTag (noise) — must not classify monorepo
  as freeHost solely for that pin dep.
- Deferred consumer peel: after pin cleaned go.mod, DIRTY-only dirt may
  soft-skip or gen-commit non-go.mod dirt; hook never sees external replace.
- `--merge-back` keeps linked monorepo worktree; `--push` publishes free when clear.
- Offline file:// modproxy supplies free old+next for tidy after tag.

## Errors

- None on success path.
- Today RED: exit ≠ 0 and/or pre-commit `local external replace forbidden` when
  monorepo peels early as false freeHost before cascade pin of free.

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
		// Today RED: false freeHost early peel gen-commits with external replace
		// still present → no-local-replace hook fails (or peel aborts).
		t.Fatalf("T-spl false freeHost intra pins replace-only: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	// Production / fixture hook surface must not appear on success.
	assertNoLocalReplaceGenCommitFail(t, out)
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("T-spl: unknown revision after apply\ncombined:\n%s", out)
	}

	// Free tagged and pushed before/with successful pin resolution.
	assertLeafMainAdvancedAndTagged(t, req)

	// Core order: cascade free tag-next before consumer pin of free @ next.
	assertFreeTagNextBeforeConsumerPinOfFree(t, out)

	// Consumer require free @ next; drop external replace.
	assertConsumerRequireAndNoExternalReplace(t, req)

	// Intra keep-local must not be force-dropped with the external pin.
	assertIntraSharedReplaceKept(t, req)

	// Pin auto-commit for free dep present (replace dropped via cascade, not feature).
	hist := historyRepoForConsumer(t, req)
	assertCascadePinCommitPresent(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertGoModCommittedClean(t, hist)

	// Replace-only: must not invent a feature gen-commit whose tree still carries
	// external replace for free (pin owns the drop). FEATURE_WIP is not seeded.
	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix+req.LeafModulePath) &&
		!strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	// If a feat: gen-commit exists, it must not reintroduce external replace.
	if featSHA := featureCommitSHA(t, hist, cascadeFeatureCommitSubject); featSHA != "" {
		featGoMod := gitOutputIsolated(t, hist, "show", featSHA+":go.mod")
		for _, line := range strings.Split(featGoMod, "\n") {
			trim := strings.TrimSpace(line)
			if !strings.Contains(trim, req.LeafModulePath) || !strings.Contains(trim, "=>") {
				continue
			}
			if strings.Contains(trim, "./external/") || strings.Contains(trim, "../") {
				t.Fatalf("feature gen-commit must not carry external replace for %s\n%s",
					req.LeafModulePath, featGoMod)
			}
		}
	}
}
```

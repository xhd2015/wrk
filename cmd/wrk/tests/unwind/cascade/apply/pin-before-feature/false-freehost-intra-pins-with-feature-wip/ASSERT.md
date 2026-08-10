## Expected

- Exit code 0.
- Free-first ship for dirty free leaf:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Order:** combined stdout/stderr shows
  `tag-next example.com/dot-pkgs @ v0.0.2` **before** consumer pin of free @
  `v0.0.2` (cascade pin after free tag; monorepo not early-peeled as false freeHost).
- Consumer monorepo after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace for that module
  - cascade pin auto-commit present for free @ next
  - pin commit is an **ancestor** of feature gen-commit (`feat: add feature`)
  - `FEATURE_WIP.md` landed via gen-commit (not scooped into pin)
  - keep intra replace `=> ./pkgs/shared`
- go.mod/go.sum clean on consumer history checkout after land.
- Combined output must **not** contain `unknown revision` or no-local-replace fail.

## Side Effects

- Early peel: free only. Monorepo pure pin-consumer deferred.
- Cascade may also pin shared @ LatestTag (noise) — must not classify monorepo
  as freeHost solely for that pin dep.
- Deferred consumer peel: feature gen-commit after pin cleaned go.mod.
- `--merge-back` keeps linked monorepo worktree; `--push` publishes free when clear.

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
		t.Fatalf("A2 false freeHost + FEATURE_WIP: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}

	assertNoLocalReplaceGenCommitFail(t, out)
	if strings.Contains(strings.ToLower(out), "unknown revision") {
		t.Fatalf("A2: unknown revision after apply\ncombined:\n%s", out)
	}

	assertLeafMainAdvancedAndTagged(t, req)
	assertFreeTagNextBeforeConsumerPinOfFree(t, out)
	assertConsumerRequireAndNoExternalReplace(t, req)
	assertIntraSharedReplaceKept(t, req)

	hist := historyRepoForConsumer(t, req)
	assertCascadePinForDepBeforeFeatureCommit(t, hist, req.LeafModulePath, req.ExpectedPinVersion, cascadeFeatureCommitSubject)
	assertPinCommitForDepFilesOnlyModSum(t, hist, req.LeafModulePath)
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("expected cascade pin commit in consumer log:\n%s", log)
	}
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("expected feature subject %q in consumer log:\n%s", cascadeFeatureCommitSubject, log)
	}
}
```

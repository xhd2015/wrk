## Expected

- Exit code 0.
- Free-first ship for dirty free leaf:
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- **Consumer root (hard lock for this leaf):**
  - Local tag `v0.0.2` exists on consumer main.
  - Tag `v0.0.2` points at consumer **main HEAD** after merge-back (not missing
    entirely; not only an intermediate cascade tip).
  - Cascade pin commit for free present on consumer history.
  - Feature WIP landed; go.mod/go.sum committed clean.
  - `require example.com/dot-pkgs v0.0.2`; no droppable external replace.

## Side Effects

- B1: early free peel → cascade tag/pin → deferred consumer peel (feature).
- `--merge-back` keeps worktrees; `--push` publishes free (and consumer) tags when clear.
- Consumer may receive cascade pin commit(s) before its own next tag.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		// Today RED may surface as exit 0 with missing consumer tag, or non-zero.
		t.Fatalf("A-wip-tag consumer-at-latest-wip: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" {
		t.Fatal("SecondRepo (free leaf) and MainRepo (consumer) required")
	}

	// Free tagged + pushed.
	assertLeafMainAdvancedAndTagged(t, req)

	// Hard lock: pure pin-consumer next tag must sit on main HEAD after full recipe.
	// Coverage gap vs T2 (pins/feature only) and A-root-tag (diamond with
	// committed owned-changed past LatestTag before cascade).
	// Observed RED: exit 0; free ships; pin + feature land; consumer tip past
	// LatestTag with **no** next tag (cascade planned while HEAD==LatestTag WIP-only).
	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("A-wip-tag: consumer root missing tag %s entirely after full B1 unwind\nlog:\n%s\ncombined:\n%s",
			unwindApplyNextTag,
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-25"),
			out)
	}
	tagSHA := revParseRef(t, req.MainRepo, "refs/tags/"+unwindApplyNextTag)
	head := revParseHEAD(t, req.MainRepo)
	if tagSHA != head {
		t.Fatalf("A-wip-tag: consumer root tag %s at %s != main HEAD %s (tagged before deferred feature peel?)\nlog:\n%s\ncombined:\n%s",
			unwindApplyNextTag, tagSHA, head,
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-25"),
			out)
	}

	// Consumer require free @ next; no external replace; feature landed.
	hist := historyRepoForConsumer(t, req)
	rootMod := filepath.Join(hist, "go.mod")
	if got := requireVersionInGoMod(t, rootMod, unwindDotPkgsModule); got != unwindApplyNextTag {
		t.Fatalf("A-wip-tag: consumer require free %s = %q, want %s\ngo.mod:\n%s",
			unwindDotPkgsModule, got, unwindApplyNextTag, readFile(t, rootMod))
	}
	if goModHasReplace(t, rootMod, unwindDotPkgsModule) {
		t.Fatalf("A-wip-tag: consumer must DROP external replace for free:\n%s", readFile(t, rootMod))
	}

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("A-wip-tag: expected cascade pin commit on consumer history:\n%s", log)
	}
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)
}
```

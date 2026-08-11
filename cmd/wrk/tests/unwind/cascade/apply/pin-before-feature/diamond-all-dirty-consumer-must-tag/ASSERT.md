## Expected

- Exit code 0.
- Free-first ship for dirty free leaf B (dot-pkgs):
  - Leaf main advanced; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` when push completes.
- Mid freeHost C (agent-pro):
  - Local tag `v0.0.2` exists on mid main (owned-changed mid).
  - `require example.com/dot-pkgs v0.0.2`; no droppable external replace for free.
- **Consumer root A (hard lock for this leaf):**
  - Local tag `v0.0.2` exists on root main.
  - Tag `v0.0.2` points at root **main HEAD** after merge-back (not only an
    intermediate tip before deferred feature gen-commit).
  - Cascade pin commit(s) for free and/or mid present on root history.
  - Feature WIP landed when present; go.mod/go.sum committed clean.
  - `require` for free and mid at `v0.0.2`; no droppable external replaces.

## Side Effects

- B1: early free/mid peels → cascade tag/pin → deferred root peel (feature).
- `--merge-back` keeps worktrees; `--push` publishes free (and other) tags when clear.
- Root may receive multiple cascade pin commits (B then C) before its own tag.

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
		// Today RED may surface as non-zero (pin/tidy) or exit 0 with missing root tag.
		t.Fatalf("A-root-tag diamond all-dirty: want exit 0; exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.DepPath == "" || req.MainRepo == "" {
		t.Fatal("SecondRepo (B), DepPath (C), MainRepo (A) required")
	}

	// B free tagged + pushed.
	assertLeafMainAdvancedAndTagged(t, req)

	// C mid: next tag present (owned-changed mid).
	if !tagRefExists(t, req.DepPath, unwindApplyNextTag) {
		t.Fatalf("A-root-tag: mid freeHost missing tag %s\nlog:\n%s",
			unwindApplyNextTag, gitOutputIsolated(t, req.DepPath, "log", "--oneline", "-20"))
	}
	assertMidPinnedToFreeLeaf(t, req)

	// Hard lock: consumer root A next tag must sit on main HEAD after full recipe.
	// Coverage gap vs T2/C-AP2/T-tag1 (pins only; no root tag-at-HEAD assert).
	// Observed RED: tag may exist from mid-cascade but tip moves after deferred
	// feature gen-commit — live symptom "A has no tag" at HEAD.
	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("A-root-tag: consumer root missing tag %s entirely after full B1 unwind\nlog:\n%s\ncombined:\n%s",
			unwindApplyNextTag,
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-25"),
			out)
	}
	tagSHA := revParseRef(t, req.MainRepo, "refs/tags/"+unwindApplyNextTag)
	head := revParseHEAD(t, req.MainRepo)
	if tagSHA != head {
		t.Fatalf("A-root-tag: consumer root tag %s at %s != main HEAD %s (tagged before deferred feature peel?)\nlog:\n%s\ncombined:\n%s",
			unwindApplyNextTag, tagSHA, head,
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-25"),
			out)
	}

	// Root pins free and/or mid @ next; no external replaces; feature landed.
	hist := historyRepoForConsumer(t, req)
	rootMod := filepath.Join(hist, "go.mod")
	if got := requireVersionInGoMod(t, rootMod, unwindDotPkgsModule); got != unwindApplyNextTag {
		t.Fatalf("A-root-tag: root require free %s = %q, want %s\ngo.mod:\n%s",
			unwindDotPkgsModule, got, unwindApplyNextTag, readFile(t, rootMod))
	}
	if got := requireVersionInGoMod(t, rootMod, unwindAgentProModule); got != unwindApplyNextTag {
		t.Fatalf("A-root-tag: root require mid %s = %q, want %s\ngo.mod:\n%s",
			unwindAgentProModule, got, unwindApplyNextTag, readFile(t, rootMod))
	}
	if goModHasReplace(t, rootMod, unwindDotPkgsModule) {
		t.Fatalf("A-root-tag: root must DROP external replace for free:\n%s", readFile(t, rootMod))
	}
	if goModHasReplace(t, rootMod, unwindAgentProModule) {
		t.Fatalf("A-root-tag: root must DROP external replace for mid:\n%s", readFile(t, rootMod))
	}

	log := gitOutputIsolated(t, hist, "log", "--oneline", "-25")
	if !strings.Contains(log, cascadePinCommitPrefix) {
		t.Fatalf("A-root-tag: expected cascade pin commit on root history:\n%s", log)
	}
	assertFeatureWIPLanded(t, hist)
	assertGoModCommittedClean(t, hist)
}
```

## Expected

- Exit code 0.
- Free-first across stack:
  - Leaf main advanced with feature content; local tag `v0.0.2` at leaf main HEAD.
  - Leaf bare origin has `main` + `v0.0.2` (leaf published when no pending leaf modules).
  - Root consumer require `example.com/dot-pkgs` bumped to `v0.0.2`.
  - Cascade pin commit on **root** history (`wrk: cascade pin …`).
  - Root go.mod/go.sum committed clean after pin (not left as uncommitted pin WIP).
- Apply peel banner for nested leaf uses relative display (`external/…`) when
  peel banners are printed (statusDirLine policy).

## Side Effects

- Land prelude merges linked leaf; cascade tags one scope on leaf, pins root,
  selective commit; push when each main has no remaining pending cascade modules.
- `--done` may remove leaf external worktree (OK).
- Keep-replace is covered by single-repo clean leaf (not this multi-repo `--done` path).

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
		t.Fatalf("apply cascade multi-repo: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.OriginBare == "" {
		t.Fatal("SecondRepo, MainRepo, OriginBare required")
	}

	// Leaf free-first ship: content + tag + origin (before/with consumer pin).
	assertLeafMainAdvancedAndTagged(t, req)

	// Banner display fidelity when peel still prints (land prelude).
	if req.DepsLinkedWtDir != "" {
		display := peelDisplay(t, req, req.DepsLinkedWtDir)
		if strings.HasPrefix(display, "external/") {
			banner := applyBannerLine(display)
			if strings.Contains(out, "==== unwind: peel") && !strings.Contains(out, banner) &&
				!strings.Contains(out, "==== unwind: peel external/") {
				t.Fatalf("when peel banners print, want relative external display %q\nout:\n%s",
					banner, out)
			}
		}
	}

	// Consumer pin bump + cascade pin commit (module cascade, not peel-only pin WIP).
	assertRequireBumped(t, req)
	assertCascadePinCommitPresent(t, req.MainRepo, unwindDotPkgsModule, unwindApplyNextTag)
	assertGoModCommittedClean(t, req.MainRepo)

	// Free-first evidence: leaf tag exists on leaf main (already asserted) and
	// root pin commit mentions leaf module — pin cannot precede leaf tag creation
	// on a correct cascade (leaf tag must exist for pin version).
	if !tagRefExists(t, req.SecondRepo, unwindApplyNextTag) {
		t.Fatalf("leaf tag %s missing — free-first cascade tags leaf before consumer pin",
			unwindApplyNextTag)
	}
}
```
